package client

import (
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/gw/application/auth"
	"github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	requestEventFactory *request.RequestEventFactory
	authHandlers        auth.Handlers
	cmapAPI             cmap.SlaveAPI
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cmapAPI cmap.SlaveAPI, f *request.RequestEventFactory, authHandlers auth.Handlers) Handlers {
	logger = mlog.GetPackageLogger("app/gw/application/client")

	return &handlers{
		requestEventFactory: f,
		authHandlers:        authHandlers,
		cmapAPI:             cmapAPI,
	}
}

func (h *handlers) getObjectLocation(oid, bucket string) (*nilrpc.MOBObjectGetResponse, error) {
	mds, err := h.cmapAPI.SearchCall().Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
	if err != nil {
		return nil, errors.Wrap(err, "find alive mds failed")
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "dial to mds failed")
	}
	defer conn.Close()

	req := &nilrpc.MOBObjectGetRequest{
		Name:   bucket + "." + strings.Replace(oid, "/", ".", -1),
		Bucket: bucket,
	}
	res := &nilrpc.MOBObjectGetResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsObjectGet.String(), req, res); err != nil {
		logger.Errorf("%+v", req)
		return nil, errors.Wrap(err, "mds rpc client calling failed")
	}

	return res, nil
}

// Handlers is the interface that provides client http handlers.
type Handlers interface {
	MakeBucketHandler(w http.ResponseWriter, r *http.Request)
	RemoveBucketHandler(w http.ResponseWriter, r *http.Request)

	PutObjectHandler(w http.ResponseWriter, r *http.Request)
	GetObjectHandler(w http.ResponseWriter, r *http.Request)
	DeleteObjectHandler(w http.ResponseWriter, r *http.Request)
}
