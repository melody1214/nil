package membership

import (
	"time"

	"github.com/chanyoung/nil/app/ds/delivery"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type handlers struct {
	cfg  *config.Ds
	cMap *cmap.CMap

	swimSrv    *swim.Server
	swimTransL *nilmux.SwimTransportLayer
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Ds) delivery.MembershipHandlers {
	log = mlog.GetLogger().WithField("package", "ds/usecase/admin")

	return &handlers{
		cfg:  cfg,
		cMap: cmap.New(),
	}
}

// Create makes swim server.
func (h *handlers) Create(swimL *nilmux.Layer) (err error) {
	h.swimTransL = nilmux.NewSwimTransportLayer(swimL)

	// Setup configuration.
	swimConf := swim.DefaultConfig()
	swimConf.ID = swim.ServerID(h.cfg.ID)
	swimConf.Address = swim.ServerAddress(h.cfg.ServerAddr + ":" + h.cfg.ServerPort)
	swimConf.Coordinator = swim.ServerAddress(h.cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(h.cfg.Swim.Period); err != nil {
		log.Error(err)
	} else {
		swimConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(h.cfg.Swim.Expire); err != nil {
		log.Error(err)
	} else {
		swimConf.PingExpire = t
	}
	swimConf.Type = swim.DS

	h.swimSrv, err = swim.NewServer(swimConf, h.swimTransL)
	if err != nil {
		log.Error(err)
		return
	}

	return nil
}

// Run starts swim service.
func (h *handlers) Run() {
	sc := make(chan swim.PingError, 1)
	go h.swimSrv.Serve(sc)

	for {
		select {
		case err := <-sc:
			log.WithFields(logrus.Fields{
				"server":       "swim",
				"message type": err.Type,
				"destID":       err.DestID,
			}).Error(err.Err)
		}
	}
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
