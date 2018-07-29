package object

import (
	"net/http"

	"github.com/chanyoung/nil/pkg/client"
	cr "github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	requestEventFactory *cr.RequestEventFactory
	// chunkPool           *chunkPool
	store Repository
	// endec               *endec
	cmapAPI cmap.SlaveAPI
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Ds, cmapAPI cmap.SlaveAPI, f *cr.RequestEventFactory, s Repository) (Handlers, error) {
	logger = mlog.GetPackageLogger("app/ds/usecase/object")

	// shards, err := strconv.ParseInt(cfg.LocalParityShards, 10, 64)
	// if err != nil {
	// 	return nil, err
	// }
	// chunkSize, err := strconv.ParseInt(cfg.ChunkSize, 10, 64)
	// if err != nil {
	// 	return nil, err
	// }

	// pool := newChunkPool(cmapAPI, shards, chunkSize, s.GetChunkHeaderSize(), s.GetObjectHeaderSize(), chunkSize-40000)

	// ed, err := newEndec(cmapAPI, pool, s)
	// if err != nil {
	// 	return nil, err
	// }
	// go ed.Run()

	return &handlers{
		requestEventFactory: f,
		// chunkPool:           pool,
		// endec:               ed,
		store:   s,
		cmapAPI: cmapAPI,
	}, nil
}

// PutObjectHandler handles the client request for creating an object.
func (h *handlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
	// ctxLogger := mlog.GetMethodLogger(logger, "handlers.PutObjectHandler")

	// req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	// if err == client.ErrInvalidProtocol {
	// 	ctxLogger.Error(errors.Wrap(err, "failed to create request event"))
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// switch req.Type() {
	// case client.WriteToPrimary:
	// 	h.writeToPrimary(req)
	// case client.WriteToFollower:
	// 	h.writeCopy(req)
	// default:
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
}

// GetChunkHandler handles the client request for downloading a chunk.
func (h *handlers) GetChunkHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
	// storeReq := &repository.Request{
	// 	Op:     repository.ReadAll,
	// 	Vol:    r.Header.Get("Volume"),
	// 	LocGid: r.Header.Get("Encoding-Group"),
	// 	Cid:    r.Header.Get("Chunk-Name"),
	// 	Osize:  h.chunkPool.chunkSize,
	// 	Out:    w,
	// }

	// shard := r.Header.Get("Shard")
	// if shard != "" {
	// 	truncateReq := &repository.Request{
	// 		Op:     repository.Write,
	// 		Vol:    r.Header.Get("Volume"),
	// 		LocGid: r.Header.Get("Encoding-Group"),
	// 		Oid:    "fake, just for truncating",
	// 		Osize:  1000000000,
	// 		Cid:    r.Header.Get("Chunk-Name") + "_" + shard,
	// 		In:     &io.PipeReader{},
	// 		Md5:    "fakemd5stringfakemd5stringfakemd",
	// 	}

	// 	c, ok := h.chunkPool.pool[chunkID(storeReq.Cid)[2:]]
	// 	if ok {
	// 		if err := h.store.Push(truncateReq); err != nil {
	// 			logger.Error(errors.Wrap(err, "failed to push truncated request"))
	// 			http.Error(w, err.Error(), http.StatusInternalServerError)
	// 			return
	// 		}
	// 		if err := truncateReq.Wait(); err == nil {
	// 			logger.Error(fmt.Errorf("truncate request returns no error"))
	// 			http.Error(w, err.Error(), http.StatusInternalServerError)
	// 			return
	// 		} else if err.Error() != "truncated" {
	// 			logger.Error(errors.Wrap(err, "failed to truncat chunk"))
	// 			http.Error(w, err.Error(), http.StatusInternalServerError)
	// 			return
	// 		}
	// 		h.chunkPool.FinishWriting(chunkID(truncateReq.Cid), h.chunkPool.chunkSize)
	// 	}

	// 	c, ok = h.chunkPool.writing[chunkID(storeReq.Cid[2:])]
	// 	if ok {
	// 		err := fmt.Errorf("try later. requested chunk is writing now")
	// 		logger.Error(err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	c, ok = h.chunkPool.encoding[chunkID(storeReq.Cid[2:])]
	// 	if ok && c.encoding {
	// 		err := fmt.Errorf("try later. requested chunk is encoding now")
	// 		logger.Error(err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	storeReq.Cid = storeReq.Cid + "_" + shard
	// }

	// h.store.Push(storeReq)
	// err := storeReq.Wait()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.WriteHeader(http.StatusOK)
}

// PutChunkHandler handles the client request for uploading a chunk.
func (h *handlers) PutChunkHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
	// storeReq := &repository.Request{
	// 	Op:     repository.WriteAll,
	// 	Vol:    r.Header.Get("Volume"),
	// 	LocGid: r.Header.Get("Encoding-Group"),
	// 	Cid:    r.Header.Get("Chunk-Name"),
	// 	Osize:  h.chunkPool.chunkSize,
	// 	In:     r.Body,
	// }

	// h.store.Push(storeReq)
	// err := storeReq.Wait()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// r.Body.Close()
	// w.WriteHeader(http.StatusOK)
}

func (h *handlers) writeToPrimary(req client.RequestEvent) {
	// ctxLogger := mlog.GetMethodLogger(logger, "handlers.writeToPrimary")

	// contentLength, err := strconv.ParseInt(req.Request().Header.Get("Content-Length"), 10, 64)
	// if err != nil {
	// 	ctxLogger.Error(errors.Wrap(err, "invalid content length"))
	// 	req.SendInternalError()
	// 	return
	// }

	// // Write into the available chunk.
	// cid, err := h.chunkPool.FindAvailableChunk(
	// 	egID(req.Request().Header.Get("Local-Chain-Id")),
	// 	vID(req.Request().Header.Get("Volume-Id")), contentLength,
	// )
	// if err != nil {
	// 	ctxLogger.Error(errors.Wrap(err, "failed to find available chunk"))
	// 	req.SendInternalError()
	// 	return
	// }

	// if len(req.MD5()) < 32 {
	// 	req.SendInternalError()
	// 	return
	// }

	// storeReq := &repository.Request{
	// 	Op:     repository.Write,
	// 	Vol:    req.Request().Header.Get("Volume-Id"),
	// 	LocGid: req.Request().Header.Get("Local-Chain-Id"),
	// 	Oid:    strings.Replace(strings.Trim(req.Request().RequestURI, "/"), "/", ".", -1),
	// 	Cid:    string(cid),
	// 	Osize:  contentLength,
	// 	Md5:    req.MD5(),

	// 	In: req.Request().Body,
	// }

	// err = h.store.Push(storeReq)
	// if err != nil {
	// 	ctxLogger.Error(errors.Wrap(err, "failed to push writing request into the backend store"))
	// 	h.chunkPool.FinishWriting(cid, 0)
	// 	req.SendInternalError()
	// 	return
	// }

	// err = storeReq.Wait()
	// if err != nil {
	// 	// TODO: handling error in writing process.

	// 	// Rollback writed data.
	// 	storeReq.Cid = ""
	// 	storeReq.Op = repository.DeleteReal
	// 	h.store.Push(storeReq)
	// 	storeReq.Wait()
	// 	h.chunkPool.FinishWriting(cid, 0)

	// 	ctxLogger.Error(errors.Wrap(err, "failed to write into the backend store"))
	// 	req.SendInternalError()
	// 	return
	// }

	// // Copy to the remote follower node
	// err = h.writeToRemoteFollower(req, contentLength, cid)
	// if err != nil {
	// 	// TODO: handling error in writing.

	// 	// Rollback writed data.
	// 	storeReq.Cid = ""
	// 	storeReq.Op = repository.DeleteReal
	// 	h.store.Push(storeReq)
	// 	storeReq.Wait()
	// 	h.chunkPool.FinishWriting(cid, 0)

	// 	ctxLogger.Error(errors.Wrap(err, "failed to write to remote follower"))
	// 	req.SendInternalError()
	// 	return
	// }

	// // Commit writed data.
	// h.chunkPool.FinishWriting(cid, contentLength)

	// req.SendSuccess()
}

// func (h *handlers) writeToRemoteFollower(req client.RequestEvent, size int64, cid chunkID) error {
// c, ok := h.chunkPool.GetChunk(cid)
// if ok == false {
// 	return fmt.Errorf("no such chunk: %s", cid)
// }

// encGrpID, err := strconv.ParseInt(string(c.encodingGroup), 10, 64)
// if err != nil {
// 	return errors.Wrap(err, "failed to convert encoding group id")
// }

// call := h.cmapAPI.SearchCall()
// encGrp, err := call.EncGrp().ID(cmap.ID(encGrpID)).Do()
// if err != nil {
// 	return errors.Wrapf(err, "failed to find such encoding group: %d", encGrpID)
// }

// vol, err := call.Volume().ID(encGrp.Vols[c.shard].ID).Do()
// if err != nil {
// 	return errors.Wrapf(err, "failed to find such volume: %d", encGrp.Vols[c.shard])
// }

// node, err := call.Node().ID(vol.Node).Do()
// if err != nil {
// 	return errors.Wrapf(err, "failed to find such node: %d", vol.Node)
// }

// remoteAddr := "https://" + node.Addr.String() + req.Request().RequestURI

// pReader, pWriter := io.Pipe()

// storeReq := &repository.Request{
// 	Op:     repository.Read,
// 	Vol:    string(c.volume),
// 	LocGid: string(c.encodingGroup),
// 	Oid:    strings.Replace(strings.Trim(req.Request().RequestURI, "/"), "/", ".", -1),
// 	Cid:    string(cid),
// 	Osize:  size,

// 	Out: pWriter,
// }

// err = h.store.Push(storeReq)
// if err != nil {
// 	return errors.Wrap(err, "failed to push read request to store")
// }

// go func(readReq *repository.Request) {
// 	defer pWriter.Close()
// 	err := readReq.Wait()
// 	if err != nil {
// 		logger.WithField("method", "handlers.writeToRemoteFollower").Errorf("failed to read from store: %+v", err)
// 		return
// 	}
// }(storeReq)

// headers := client.NewHeaders()
// headers.SetLocalChainID(encGrp.ID.String())
// headers.SetVolumeID(vol.ID.String())
// headers.SetChunkID(c.status.String() + "_" + string(c.id))
// headers.SetMD5(req.MD5())

// copyReq, err := cr.NewRequest(
// 	client.WriteToFollower, req.Request().Method,
// 	remoteAddr, pReader, headers, size,
// 	cr.WithS3(true), cr.WithCopyHeaders(req.CopyAuthHeader()),
// )
// if err != nil {
// 	return errors.Wrap(err, "failed to create copy request")
// }

// resp, err := copyReq.Send()
// if err != nil {
// 	return errors.Wrap(err, "failed to send copy request")
// }

// if resp.StatusCode != http.StatusOK {
// 	return fmt.Errorf("copy request returns http status code: %d", resp.StatusCode)
// }

// return nil
// }

// // writeCopy writes the copy request from the primary into the store.
// func (h *handlers) writeCopy(req client.RequestEvent) {
// ctxLogger := mlog.GetMethodLogger(logger, "handlers.writeCopy")

// contentLength, err := strconv.ParseInt(req.Request().Header.Get("Content-Length"), 10, 64)
// if err != nil {
// 	ctxLogger.Error(errors.Wrap(err, "failed to parse object size"))
// 	req.SendInternalError()
// 	return
// }

// tmp, err := strconv.ParseInt(req.Request().Header.Get("Volume-Id"), 10, 64)
// if err != nil {
// 	ctxLogger.Error(errors.Wrap(err, "failed to parse volume id"))
// 	req.SendInternalError()
// 	return
// }
// vol := cmap.ID(tmp)

// tmp, err = strconv.ParseInt(req.Request().Header.Get("Local-Chain-Id"), 10, 64)
// if err != nil {
// 	ctxLogger.Error(errors.Wrap(err, "failed to parse encoding group id"))
// 	req.SendInternalError()
// 	return
// }
// encgrp := cmap.ID(tmp)

// if len(req.MD5()) < 32 {
// 	req.SendInternalError()
// 	return
// }

// storeReq := &repository.Request{
// 	Op:     repository.Write,
// 	Vol:    vol.String(),
// 	Oid:    strings.Replace(strings.Trim(req.Request().URL.Path, "/"), "/", ".", -1),
// 	Cid:    req.Request().Header.Get("Chunk-Id"),
// 	LocGid: encgrp.String(),
// 	Osize:  contentLength,
// 	Md5:    req.MD5(),

// 	In: req.Request().Body,
// }

// err = h.store.Push(storeReq)
// if err != nil {
// 	ctxLogger.Error(errors.Wrap(err, "failed to push writing request into the backend store"))
// 	req.SendInternalError()
// 	return
// }

// err = storeReq.Wait()
// if err != nil {
// 	// TODO: handling error in writing process.

// 	// Rollback writed data.
// 	storeReq.Cid = ""
// 	storeReq.Op = repository.DeleteReal
// 	h.store.Push(storeReq)
// 	storeReq.Wait()

// 	ctxLogger.Error(errors.Wrap(err, "failed to write into the backend store"))
// 	req.SendInternalError()
// 	return
// }

// mds, err := h.cmapAPI.SearchCall().Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
// if err != nil {
// 	// Rollback writed data.
// 	storeReq.Cid = ""
// 	storeReq.Op = repository.DeleteReal
// 	h.store.Push(storeReq)
// 	storeReq.Wait()

// 	ctxLogger.Error(errors.Wrap(err, "failed to get alive mds from cmap map"))
// 	req.SendInternalError()
// 	return
// }

// conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
// if err != nil {
// 	// Rollback writed data.
// 	storeReq.Cid = ""
// 	storeReq.Op = repository.DeleteReal
// 	h.store.Push(storeReq)
// 	storeReq.Wait()

// 	ctxLogger.Error(errors.Wrap(err, "failed to dial to mds"))
// 	req.SendInternalError()
// 	return
// }
// defer conn.Close()

// metaReq := &nilrpc.MOBObjectPutRequest{
// 	Name:          storeReq.Oid,
// 	Bucket:        strings.Split(strings.Trim(req.Request().URL.Path, "/"), "/")[0],
// 	EncodingGroup: encgrp,
// 	Volume:        vol,
// }
// metaRes := &nilrpc.MOBObjectPutResponse{}

// cli := rpc.NewClient(conn)
// if err := cli.Call(nilrpc.MdsObjectPut.String(), metaReq, metaRes); err != nil {
// 	// Rollback writed data.
// 	storeReq.Cid = ""
// 	storeReq.Op = repository.DeleteReal
// 	h.store.Push(storeReq)
// 	storeReq.Wait()

// 	ctxLogger.Error(errors.Wrap(err, "failed to write object meta to the mds"))
// 	req.SendInternalError()
// 	return
// }

// req.SendSuccess()
// 	http.Error(w, "not implemented", http.StatusNotImplemented)
// }

// GetObjectHandler handles the client request for getting an object.
func (h *handlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	// ctxLogger := mlog.GetMethodLogger(logger, "handlers.GetObjectHandler")

	// _, err := h.requestEventFactory.CreateRequestEvent(w, r)
	// if err == client.ErrInvalidProtocol {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// size, ok := h.store.GetObjectSize(r.Header.Get("Volume-Id"), strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1))
	// if ok == false {
	// 	ctxLogger.Error("failed to get object size")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	// md5, ok := h.store.GetObjectMD5(r.Header.Get("Volume-Id"), strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1))
	// if ok == false {
	// 	ctxLogger.Error("failed to get object md5")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("ETag", md5)

	// storeReq := &repository.Request{
	// 	Op:     repository.Read,
	// 	Vol:    r.Header.Get("Volume-Id"),
	// 	Oid:    strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1),
	// 	LocGid: r.Header.Get("Local-Chain-Id"),
	// 	Osize:  size,

	// 	Out: w,
	// }
	// h.store.Push(storeReq)

	// err = storeReq.Wait()
	// if err != nil {
	// 	ctxLogger.Error(errors.Wrap(err, "failed to read object"))
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// req.SendSuccess()
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// DeleteObjectHandler handles the client request for deleting an object.
func (h *handlers) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	// _, err := h.requestEventFactory.CreateRequestEvent(w, r)
	// if err == client.ErrInvalidProtocol {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// }

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *handlers) SetChunkPool(req *nilrpc.DOBSetChunkPoolRequest, res *nilrpc.DOBSetChunkPoolResponse) error {
	return nil
	// return h.chunkPool.moveChunk(
	// 	egID(req.EG.String()),
	// 	vID(req.Vol.String()),
	// 	chunkID(req.ID),
	// 	req.Shard,
	// )
}

// Handlers is the interface that provides client http handlers.
type Handlers interface {
	PutObjectHandler(w http.ResponseWriter, r *http.Request)
	GetObjectHandler(w http.ResponseWriter, r *http.Request)
	DeleteObjectHandler(w http.ResponseWriter, r *http.Request)
	GetChunkHandler(w http.ResponseWriter, r *http.Request)
	PutChunkHandler(w http.ResponseWriter, r *http.Request)
	RPCHandler() RPCHandler
}

// RPCHandler returns the RPC handler which will handle
// the requests from the delivery layer.
func (h *handlers) RPCHandler() RPCHandler {
	// This is a trick to hide inadvertently exposed methods,
	// such as Join() or Leave().
	type handler struct{ RPCHandler }
	return handler{RPCHandler: h}
}

// RPCHandler is the interface that provides clustermap domain's rpc handlers.
type RPCHandler interface {
	SetChunkPool(*nilrpc.DOBSetChunkPoolRequest, *nilrpc.DOBSetChunkPoolResponse) error
}
