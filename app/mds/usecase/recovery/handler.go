package recovery

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	worker *worker
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, cMap *cmap.Controller, store Repository) (Handlers, error) {
	logger = mlog.GetPackageLogger("app/mds/usecase/recovery")

	worker, err := newWorker(cfg, cMap, store)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create recovery worker")
	}
	go worker.run()

	return &handlers{worker: worker}, nil
}

func (h *handlers) Recover(req *nilrpc.RecoverRequest, res *nilrpc.RecoverResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Recover")
	// Logging the error.
	ctxLogger.WithFields(logrus.Fields{
		"server":       "swim",
		"message type": req.Pe.Type,
		"destID":       req.Pe.DestID,
	}).Warn(req.Pe.Err)

	go func() {
		select {
		case h.worker.recoveryCh <- nil:
			return
		case <-time.After(10 * time.Millisecond):
			return
		}
	}()

	return nil
}

func (h *handlers) Rebalance(req *nilrpc.RebalanceRequest, res *nilrpc.RebalanceResponse) error {
	go func() {
		select {
		case h.worker.rebalanceCh <- nil:
			return
		case <-time.After(10 * time.Millisecond):
			return
		}
	}()

	return nil
}

// Handlers is the interface that provides recovery domain's rpc handlers.
type Handlers interface {
	Recover(req *nilrpc.RecoverRequest, res *nilrpc.RecoverResponse) error
	Rebalance(req *nilrpc.RebalanceRequest, res *nilrpc.RebalanceResponse) error
}
