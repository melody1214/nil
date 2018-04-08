package object

import (
	"net/http"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

// PutObjectHandler handles the client request for creating an object.
func (h *handlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.PutObjectHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		// TODO: make own protocol
		osize, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
			return
		}

		mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		req := &nilrpc.ObjectPutRequest{
			Name:                storeReq.Oid,
			Bucket:              strings.Split(strings.Trim(r.URL.Path, "/"), "/")[0],
			EncodingGroup:       storeReq.LocGid,
			EncodingGroupVolume: r.Header.Get("Encoding-Group-Volume"),
		}
		res := &nilrpc.ObjectPutResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.MdsObjectPut.String(), req, res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	attrs := r.Header.Get("X-Amz-Meta-S3cmd-Attrs")

	var md5str string
	for _, attr := range strings.Split(attrs, "/") {
		if strings.HasPrefix(attr, "md5:") {
			md5str = strings.Split(attr, ":")[1]
			break
		}
	}

	encReq := newRequest(r)
	encReq.md5 = md5str
	h.encoder.Push(encReq)
	if err := encReq.wait(); err != nil {
		ctxLogger.Error(err)
		req.SendInternalError()
		return
	}

	req.ResponseWriter().Header().Set("ETag", md5str)
	req.SendSuccess()
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
		ctxLogger.Error(errors.Wrap(err, "failed to get object size"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	md5, ok := h.store.GetObjectMD5(r.Header.Get("Volume-Id"), strings.Replace(strings.Trim(r.URL.Path, "/"), "/", ".", -1))
	if ok == false {
		ctxLogger.Error(errors.Wrap(err, "failed to get object md5"))
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
