package client

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
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

	eg, node, err := h.findRandomPlaceToWrite()
	if err != nil {
		ctxLogger.Error(err)
		req.SendInternalError()
		return
	}

	rpURL, err := url.Parse("https://" + node.Addr.String())
	if err != nil {
		ctxLogger.Error(errors.Wrapf(err,
			"parse ds url failed, ds ID: %s, ds url: %s",
			node.ID.String(), node.Addr),
		)
		req.SendInternalError()
		return
	}

	const primary = 0
	proxy := httputil.NewSingleHostReverseProxy(rpURL)
	r.Header.Add("Volume-Id", eg.Vols[primary].ID.String())
	r.Header.Add("Local-Chain-Id", eg.ID.String())
	r.Header.Add("Request-Type", client.WriteToPrimary.String())
	proxy.ErrorLog = log.New(logger.Writer(), "http reverse proxy", log.Lshortfile)
	proxy.ServeHTTP(w, r)
}

func (h *handlers) findRandomPlaceToWrite() (cmap.EncodingGroup, cmap.Node, error) {
	c := h.cmapAPI.SearchCall()
	eg, err := c.EncGrp().Random().Status(cmap.EGAlive).Do()
	if err != nil {
		return cmap.EncodingGroup{}, cmap.Node{}, errors.Wrap(err, "failed to search writable encoding group")
	}

	const primary = 0
	vol, err := c.Volume().ID(eg.Vols[primary].ID).Status(cmap.VolActive).Do()
	if err != nil {
		return cmap.EncodingGroup{}, cmap.Node{}, errors.Wrapf(err, "failed to search active volume %+v", eg.Vols[primary])
	}

	node, err := c.Node().ID(vol.Node).Status(cmap.NodeAlive).Do()
	if err != nil {
		return cmap.EncodingGroup{}, cmap.Node{}, errors.Wrapf(err, "failed to search alive node %+v", vol.Node)
	}

	return eg, node, nil
}

// GetObjectHandler handles the client request for getting an object.
func (h *handlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.GetObjectHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bucketAndObject := strings.SplitN(strings.Trim(r.RequestURI, "/"), "/", 2)
	if len(bucketAndObject) < 2 {
		ctxLogger.Error(err)
		req.SendInvalidURI()
		return
	}

	res, err := h.getObjectLocation(bucketAndObject[1], bucketAndObject[0])
	if err != nil {
		ctxLogger.Error(err)
		req.SendInvalidURI()
		return
	}

	node, err := h.cmapAPI.SearchCall().Node().ID(cmap.ID(res.DsID)).Do()
	if err != nil {
		ctxLogger.Error(err)
		req.SendInternalError()
		return
	}

	rpURL, err := url.Parse("https://" + node.Addr.String())
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
	r.Header.Add("Volume-Id", res.VolumeID.String())
	r.Header.Add("Local-Chain-Id", res.EncodingGroupID.String())
	proxy.ErrorLog = log.New(logger.Writer(), "http reverse proxy", log.Lshortfile)
	proxy.ServeHTTP(w, r)
}

// DeleteObjectHandler handles the client request for deleting an object.
func (h *handlers) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}
