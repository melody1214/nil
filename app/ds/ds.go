package ds

import (
	"errors"

	"github.com/chanyoung/nil/app/ds/server"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Ds is the object storage daemon.
type Ds struct {
	// Unique id of ds.
	id string

	cfg    *config.Ds
	server *server.Server
}

// New creates a ds object.
func New(cfg *config.Ds) (*Ds, error) {
	// Setting logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return nil, err
	}

	// Get initialized logger.
	log = mlog.GetLogger()
	if log == nil {
		return nil, errors.New("nil logger object")
	}
	log.WithField("location", cfg.LogLocation).Info("Setting logger succeeded")

	// Generate DS ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating DS UUID succeeded")

	s, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	o := &Ds{
		id:     cfg.ID,
		cfg:    cfg,
		server: s,
	}

	return o, nil
}

// Start starts the ds.
func (d *Ds) Start() {
	log.Info("Start DS service ...")
	err := d.server.Start()
	if err != nil {
		log.Error(err)
	}
}

func (d *Ds) stop() {
	// Clean up ds works.
	log.Info("Stop Ds")
}
