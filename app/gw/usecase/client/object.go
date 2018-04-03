package client

import (
	golog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

func (h *handlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := log.WithField("method", "handlers.PutObjectHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	res, err := h.getLocalChain()
	if err != nil {
		ctxLogger.Error(err)
		req.SendInternalError()
		return
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

	// Test code
	c := h.cMap.SearchCall()
	node, err := c.ID(cmap.ID(res.ParityNodeID)).Do()
	if err != nil {
		ctxLogger.Error(err)
		req.SendInternalError()
		return
	}

	rpURL, err := url.Parse("https://" + node.Addr)
	if err != nil {
		ctxLogger.Error(
			errors.Wrapf(
				err,
				"parse ds url failed, ds ID: %s, ds url: %s",
				node.ID.String(),
				node.Addr,
			),
		)
		req.SendInternalError()
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(rpURL)
	r.Header.Add("Volume-Id", strconv.FormatInt(res.ParityVolumeID, 10))
	r.Header.Add("Local-Chain-Id", strconv.FormatInt(res.LocalChainID, 10))
	proxy.ErrorLog = golog.New(log.Writer(), "http reverse proxy", golog.Lshortfile)
	proxy.ServeHTTP(w, r)
}

func (h *handlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *handlers) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}
