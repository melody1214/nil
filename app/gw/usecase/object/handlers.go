package object

import (
	golog "log"
	"net/http"
	"net/http/httputil"
	"net/rpc"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/gw/delivery"
	"github.com/chanyoung/nil/app/gw/usecase/auth"
	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type handlers struct {
	requestEventFactory *request.RequestEventFactory
	authHandlers        auth.Handlers
	cMap                *cmap.CMap
}

func NewObjectHandlers(f *request.RequestEventFactory, authHandlers auth.Handlers) delivery.ObjectHandlers {
	log = mlog.GetLogger().WithField("package", "gw/usecase/object")

	return &handlers{
		requestEventFactory: f,
		authHandlers:        authHandlers,
		cMap:                cmap.New(),
	}
}

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

func (h *handlers) getLocalChain() (*nilrpc.GetLocalChainResponse, error) {
	mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			return nil, errors.Wrap(err, "find alive mds failed")
		}
	}

	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "dial to mds failed")
	}
	defer conn.Close()

	req := &nilrpc.GetLocalChainRequest{}
	res := &nilrpc.GetLocalChainResponse{}

	// 4. Request the secret key.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.GetLocalChain.String(), req, res); err != nil {
		return nil, errors.Wrap(err, "mds rpc client calling failed")
	}

	return res, nil
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := log.WithField("method", "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	h.cMap = m
}
