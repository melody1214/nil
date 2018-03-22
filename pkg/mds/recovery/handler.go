package recovery

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Handler has exposed methods for rpc server.
type Handler struct {
	store   *store.Store
	swimSrv *swim.Server
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

func (h *Handler) Recover(pe swim.PingError) {
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
