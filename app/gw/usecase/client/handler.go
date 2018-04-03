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

var log *logrus.Entry

type handlers struct {
	requestEventFactory *request.RequestEventFactory
	authHandlers        auth.Handlers
	cMap                *cmap.CMap
}

func NewHandlers(f *request.RequestEventFactory, authHandlers auth.Handlers) delivery.ClientHandlers {
	log = mlog.GetLogger().WithField("package", "gw/usecase/client")

	return &handlers{
		requestEventFactory: f,
		authHandlers:        authHandlers,
		cMap:                cmap.New(),
	}
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
