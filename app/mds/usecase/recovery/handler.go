package recovery

import (
	"log"
	"net/rpc"
	"sync"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg   *config.Mds
	store Repository
	cMap  *cmap.Controller

	l sync.RWMutex
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, cMap *cmap.Controller, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/recovery")

	return &handlers{
		cfg:   cfg,
		store: s,
		cMap:  cMap,
	}
}

func (h *handlers) Recover(req *nilrpc.RecoverRequest, res *nilrpc.RecoverResponse) error {
	h.l.Lock()
	defer h.l.Unlock()

	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Recover")

	// Logging the error.
	ctxLogger.WithFields(logrus.Fields{
		"server":       "swim",
		"message type": req.Pe.Type,
		"destID":       req.Pe.DestID,
	}).Warn(req.Pe.Err)

	// Updates membership.
	h.updateMembership()

	// Get the new version of cluster map.
	if err := h.updateClusterMap(); err != nil {
		ctxLogger.Error(err)
	}

	// If the error message is occured because just simple membership
	// changed, then finish the recover routine here.
	if req.Pe.Err == swim.ErrChanged.Error() {
		return nil
	}

	// TODO: recovery routine.
	return nil
}

func (h *handlers) Rebalance(req *nilrpc.RebalanceRequest, res *nilrpc.RebalanceResponse) error {
	h.l.Lock()
	defer h.l.Unlock()

	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Rebalance")

	if !h.needRebalance() {
		ctxLogger.Info("no need rebalance")
		return nil
	}

	ctxLogger.Info("do rebalance")
	if err := h.rebalance(); err != nil {
		ctxLogger.Error(err)
	}

	return nil
}

// Handlers is the interface that provides recovery domain's rpc handlers.
type Handlers interface {
	Recover(req *nilrpc.RecoverRequest, res *nilrpc.RecoverResponse) error
	Rebalance(req *nilrpc.RebalanceRequest, res *nilrpc.RebalanceResponse) error
}

func (h *handlers) updateClusterMap() error {
	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MCLUpdateClusterMapRequest{}
	res := &nilrpc.MCLUpdateClusterMapResponse{}

	cli := rpc.NewClient(conn)
	defer cli.Close()

	return cli.Call(nilrpc.MdsClustermapUpdateClusterMap.String(), req, res)
}
