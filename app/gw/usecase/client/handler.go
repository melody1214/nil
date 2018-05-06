package client

import (
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/gw/usecase/auth"
	"github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cluster"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	requestEventFactory *request.RequestEventFactory
	authHandlers        auth.Handlers
	clusterAPI          cluster.SlaveAPI
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(clusterAPI cluster.SlaveAPI, f *request.RequestEventFactory, authHandlers auth.Handlers) Handlers {
	logger = mlog.GetPackageLogger("app/gw/usecase/client")

	return &handlers{
		requestEventFactory: f,
		authHandlers:        authHandlers,
		clusterAPI:          clusterAPI,
	}
}

func (h *handlers) getLocalChain() (*nilrpc.GetLocalChainResponse, error) {
	mds, err := h.clusterAPI.SearchCallNode().Type(cluster.MDS).Status(cluster.Alive).Do()
	if err != nil {
		return nil, errors.Wrap(err, "find alive mds failed")
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "dial to mds failed")
	}
	defer conn.Close()

	req := &nilrpc.GetLocalChainRequest{}
	res := &nilrpc.GetLocalChainResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsAdminGetLocalChain.String(), req, res); err != nil {
		return nil, errors.Wrap(err, "mds rpc client calling failed")
	}

	return res, nil
}

func (h *handlers) getObjectLocation(oid, bucket string) (*nilrpc.ObjectGetResponse, error) {
	mds, err := h.clusterAPI.SearchCallNode().Type(cluster.MDS).Status(cluster.Alive).Do()
	if err != nil {
		return nil, errors.Wrap(err, "find alive mds failed")
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "dial to mds failed")
	}
	defer conn.Close()

	req := &nilrpc.ObjectGetRequest{
		Name:   bucket + "." + strings.Replace(oid, "/", ".", -1),
		Bucket: bucket,
	}
	res := &nilrpc.ObjectGetResponse{}

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
