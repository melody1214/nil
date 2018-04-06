package object

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/util/mlog"
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

			In: r.Body,
		}
		h.store.Push(storeReq)

		err = storeReq.Wait()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	attrs := r.Header.Get("X-Amz-Meta-S3cmd-Attrs")

	encReq := newRequest(r)
	h.encoder.Push(encReq)
	if err := encReq.wait(); err != nil {
		ctxLogger.Error(err)
		req.SendInternalError()
		return
	}

	var md5str string
	for _, attr := range strings.Split(attrs, "/") {
		if strings.HasPrefix(attr, "md5:") {
			md5str = strings.Split(attr, ":")[1]
			break
		}
	}

	req.ResponseWriter().Header().Set("ETag", md5str)
	req.SendSuccess()
}

// GetObjectHandler handles the client request for getting an object.
func (h *handlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// DeleteObjectHandler handles the client request for deleting an object.
func (h *handlers) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}
