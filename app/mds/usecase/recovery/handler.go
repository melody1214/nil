package recovery

import (
	"fmt"
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

// Recovery recieves the recovery requests from other domains.
func (h *handlers) Recovery(req *nilrpc.MRERecoveryRequest, res *nilrpc.MRERecoveryResponse) error {
	// Select the channel to send notification by the type of recovery request.
	var notiCh chan interface{}
	switch req.Type {
	case nilrpc.Recover:
		notiCh = h.worker.recoveryCh
	case nilrpc.Rebalance:
		notiCh = h.worker.rebalanceCh
	default:
		return fmt.Errorf("unknown recovery type: %+v", req.Type)
	}

	// Send the notification.
	// Expire after very short time because blocked channel means
	// the channel already pending the request.
	select {
	case notiCh <- nil:
		return nil
	case <-time.After(0):
		return nil
	}
}

// Handlers is the interface that provides recovery domain's rpc handlers.
type Handlers interface {
	Recovery(req *nilrpc.MRERecoveryRequest, res *nilrpc.MRERecoveryResponse) error
}
