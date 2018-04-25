package object

import (
	"net/http"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/client"
	cr "github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	requestEventFactory *cr.RequestEventFactory
	chunkPool           *chunkPool
	store               Repository
	endec               *endec
	cMap                *cmap.Controller
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Ds, cMap *cmap.Controller, f *cr.RequestEventFactory, s Repository) (Handlers, error) {
	logger = mlog.GetPackageLogger("app/ds/usecase/object")

	ed, err := newEncoder(cMap, s)
	if err != nil {
		return nil, err
	}
	go ed.Run()

	shards, err := strconv.ParseInt(cfg.LocalParityShards, 10, 64)
	if err != nil {
		return nil, err
	}
	chunkSize, err := strconv.ParseInt(cfg.ChunkSize, 10, 64)
	if err != nil {
		return nil, err
	}

	return &handlers{
		requestEventFactory: f,
		chunkPool:           newChunkPool(shards, chunkSize, s.GetChunkHeaderSize(), s.GetObjectHeaderSize(), chunkSize-1000),
		endec:               ed,
		store:               s,
		cMap:                cMap,
	}, nil
}

// PutObjectHandler handles the client request for creating an object.
func (h *handlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.PutObjectHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch req.Type() {
	case client.WriteToPrimary:
		h.writeToPrimary(req)
	case client.WriteToFollower:
		osize, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ctxLogger.Error(errors.Wrap(err, "failed to parse object size"))
			return
		}

		storeReq := &repository.Request{
			Op:     repository.Write,
			Vol:    r.Header.Get("Volume-Id"),
			Oid:    strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1),
			Cid:    r.Header.Get("Chunk-Id"),
			LocGid: r.Header.Get("Local-Chain-Id"),
			Osize:  osize,
			Md5:    r.Header.Get("Md5"),

			In: r.Body,
		}
		h.store.Push(storeReq)

		err = storeReq.Wait()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ctxLogger.Error(errors.Wrap(err, "failed to wait backend store request finish"))
			return
		}

		mds, err := h.cMap.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ctxLogger.Error(errors.Wrap(err, "failed to get alive mds from cluster map"))
			return
		}

		conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ctxLogger.Error(errors.Wrap(err, "failed to dial to mds"))
			return
		}
		defer conn.Close()

		req := &nilrpc.ObjectPutRequest{
			Name:          storeReq.Oid,
			Bucket:        strings.Split(strings.Trim(r.URL.Path, "/"), "/")[0],
			EncodingGroup: storeReq.LocGid,
			Volume:        r.Header.Get("Volume-Id"),
		}
		res := &nilrpc.ObjectPutResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.MdsObjectPut.String(), req, res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ctxLogger.Error(errors.Wrap(err, "failed to write object meta to the mds"))
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *handlers) writeToPrimary(req client.RequestEvent) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.writeToPrimary")

	contentLength, err := strconv.ParseInt(req.Request().Header.Get("Content-Length"), 10, 64)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "invalid content length"))
		req.SendInternalError()
		return
	}

	// Write into the available chunk.
	cid := h.chunkPool.FindAvailableChunk(
		egID(req.Request().Header.Get("Local-Chain-Id")),
		vID(req.Request().Header.Get("Volume-Id")), contentLength,
	)

	storeReq := &repository.Request{
		Op:     repository.Write,
		Vol:    req.Request().Header.Get("Volume-Id"),
		LocGid: req.Request().Header.Get("Local-Chain-Id"),
		Oid:    strings.Replace(strings.Trim(req.Request().RequestURI, "/"), "/", ".", -1),
		Cid:    string(cid),
		Osize:  contentLength,

		In: req.Request().Body,
	}

	err = h.store.Push(storeReq)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to push writing request into the backend store"))
		req.SendInternalError()
		return
	}

	err = storeReq.Wait()
	if err != nil {
		// TODO: handling error in writing process.

		// Rollback writed data.
		storeReq.Op = repository.Delete
		h.store.Push(storeReq)
		storeReq.Wait()

		ctxLogger.Error(errors.Wrap(err, "failed to write into the backend store"))
		req.SendInternalError()
		return
	}

	// Copy to the remote follower node
	err = h.writeToRemoteFollower(req, cid)
	if err != nil {
		// TODO: handling error in writing.

		// Rollback writed data.
		storeReq.Op = repository.Delete
		h.store.Push(storeReq)
		storeReq.Wait()

		ctxLogger.Error(errors.Wrap(err, "failed to write to available chunk"))
		req.SendInternalError()
		return
	}

	// Commit writed data.
	h.chunkPool.FinishWriting(cid, contentLength)

	req.SendSuccess()
}

func (h *handlers) writeToRemoteFollower(req client.RequestEvent, cid chunkID) error {
	return nil
}

// GetObjectHandler handles the client request for getting an object.
func (h *handlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.GetObjectHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	size, ok := h.store.GetObjectSize(r.Header.Get("Volume-Id"), strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1))
	if ok == false {
		ctxLogger.Error("failed to get object size")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	md5, ok := h.store.GetObjectMD5(r.Header.Get("Volume-Id"), strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1))
	if ok == false {
		ctxLogger.Error("failed to get object md5")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("ETag", md5)

	storeReq := &repository.Request{
		Op:     repository.Read,
		Vol:    r.Header.Get("Volume-Id"),
		Oid:    strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1),
		LocGid: r.Header.Get("Local-Chain-Id"),

		Out: w,
	}
	h.store.Push(storeReq)

	err = storeReq.Wait()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to read object"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req.SendSuccess()
}

// DeleteObjectHandler handles the client request for deleting an object.
func (h *handlers) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// Handlers is the interface that provides client http handlers.
type Handlers interface {
	PutObjectHandler(w http.ResponseWriter, r *http.Request)
	GetObjectHandler(w http.ResponseWriter, r *http.Request)
	DeleteObjectHandler(w http.ResponseWriter, r *http.Request)
}
