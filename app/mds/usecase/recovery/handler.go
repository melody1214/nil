package recovery

import (
	"sync"

	"github.com/chanyoung/nil/app/mds/delivery"
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
	cMap  *cmap.CMap

	l sync.RWMutex
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, s Repository) delivery.RecoveryHandlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/recovery")

	return &handlers{
		cfg:   cfg,
		store: s,
		cMap:  cmap.New(),
	}
}

func (h *handlers) Recover(req *nilrpc.RecoverRequest, res *nilrpc.RecoverResponse) error {
	h.l.Lock()
	defer h.l.Unlock()

	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Recover")

	// 1. Logging the error.
	ctxLogger.WithFields(logrus.Fields{
		"server":       "swim",
		"message type": req.Pe.Type,
		"destID":       req.Pe.DestID,
	}).Warn(req.Pe.Err)

	// 2. Updates membership.
	h.updateMembership()

	// 3. Get the new version of cluster map.
	newCMap, err := h.updateClusterMap()
	if err != nil {
		ctxLogger.Error(err)
	}

	// 4. Save the new cluster map.
	err = newCMap.Save()
	if err != nil {
		ctxLogger.Error(err)
	}

	// 5. If the error message is occured because just simple membership
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
