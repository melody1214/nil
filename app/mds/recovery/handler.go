package recovery

import (
	"fmt"
	"sync"

	"github.com/chanyoung/nil/app/mds/store"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Handler has exposed methods for rpc server.
type Handler struct {
	store   *store.Store
	swimSrv *swim.Server

	l sync.RWMutex
}

// NewHandler returns a new membership handler.
func NewHandler(store *store.Store, swimSrv *swim.Server) (*Handler, error) {
	log = mlog.GetLogger()

	if store == nil || swimSrv == nil {
		return nil, fmt.Errorf("invalid arguments")
	}

	return &Handler{
		store:   store,
		swimSrv: swimSrv,
	}, nil
}

// Recover handle the error message from the membership protocol.
func (h *Handler) Recover(pe swim.PingError) {
	h.l.Lock()
	defer h.l.Unlock()

	// 1. Logging the error.
	log.WithFields(logrus.Fields{
		"server":       "swim",
		"message type": pe.Type,
		"destID":       pe.DestID,
	}).Warn(pe.Err)

	// 2. Updates membership.
	h.updateMembership()

	// 3. Get the new version of cluster map.
	newCMap, err := h.updateClusterMap()
	if err != nil {
		log.Error(err)
	}

	// 4. Save the new cluster map.
	err = newCMap.Save()
	if err != nil {
		log.Error(err)
	}

	// 5. If the error message is occured because just simple membership
	// changed, then finish the recover routine here.
	if pe.Err == swim.ErrChanged {
		return
	}

	// TODO: recovery routine.
}

// Rebalance checks and recalculates the local chain balance.
func (h *Handler) Rebalance() {
	h.l.Lock()
	defer h.l.Unlock()

	if !h.needRebalance() {
		log.Info("no need rebalance")
		return
	}

	log.Info("do rebalance")
	if err := h.rebalance(); err != nil {
		log.Error(err)
	}

	return
}
