package client

import (
	"net/http"
	"strings"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

// PutObjectHandler handles the client request for creating an object.
func (h *handlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.PutObjectHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// Extract bucket name and object name.
	// ex) /bucketname/object1
	// ->  bucketname/object1
	// ->  bucketName: bucketname
	// ->  objectName: object1
	bucketAndObject := strings.SplitN(strings.Trim(r.RequestURI, "/"), "/", 2)
	if len(bucketAndObject) < 2 {
		ctxLogger.Error(err)
		req.SendInvalidURI()
		return
	}

	// http.Error(w, "not implemented", http.StatusNotImplemented)
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
