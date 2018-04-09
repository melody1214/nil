package membership

import (
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg *config.Ds

	swimSrv    *swim.Server
	swimTransL *nilmux.SwimTransportLayer
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Ds) Handlers {
	logger = mlog.GetPackageLogger("app/ds/usecase/membership")

	return &handlers{
		cfg: cfg,
	}
}

// Create makes swim server.
func (h *handlers) Create(swimL *nilmux.Layer) (err error) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Create")

	h.swimTransL = nilmux.NewSwimTransportLayer(swimL)

	// Setup configuration.
	swimConf := swim.DefaultConfig()
	swimConf.ID = swim.ServerID(h.cfg.ID)
	swimConf.Address = swim.ServerAddress(h.cfg.ServerAddr + ":" + h.cfg.ServerPort)
	swimConf.Coordinator = swim.ServerAddress(h.cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(h.cfg.Swim.Period); err != nil {
		ctxLogger.Error(err)
	} else {
		swimConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(h.cfg.Swim.Expire); err != nil {
		ctxLogger.Error(err)
	} else {
		swimConf.PingExpire = t
	}
	swimConf.Type = swim.DS

	h.swimSrv, err = swim.NewServer(swimConf, h.swimTransL)
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	return nil
}

// Run starts swim service.
func (h *handlers) Run() {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Run")

	sc := make(chan swim.PingError, 1)
	go h.swimSrv.Serve(sc)

	for {
		select {
		case err := <-sc:
			ctxLogger.WithFields(logrus.Fields{
				"server":       "swim",
				"message type": err.Type,
				"destID":       err.DestID,
			}).Error(err.Err)
		}
	}
}

// Handlers is the interface that provides client http handlers.
type Handlers interface {
	Create(swimL *nilmux.Layer) error
	Run()
}
