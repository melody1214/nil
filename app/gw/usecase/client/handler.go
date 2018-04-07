package client

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/gw/delivery"
	"github.com/chanyoung/nil/app/gw/usecase/auth"
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
	cMap                *cmap.Controller
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cMap *cmap.Controller, f *request.RequestEventFactory, authHandlers auth.Handlers) delivery.ClientHandlers {
	logger = mlog.GetPackageLogger("app/gw/usecase/client")

	return &handlers{
		requestEventFactory: f,
		authHandlers:        authHandlers,
		cMap:                cMap,
	}
}

func (h *handlers) getLocalChain() (*nilrpc.GetLocalChainResponse, error) {
	mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		return nil, errors.Wrap(err, "find alive mds failed")
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
	if err := cli.Call(nilrpc.MdsAdminGetLocalChain.String(), req, res); err != nil {
		return nil, errors.Wrap(err, "mds rpc client calling failed")
	}

	return res, nil
}
