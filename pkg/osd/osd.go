package osd

import (
	"errors"

	"github.com/chanyoung/nil/pkg/osd/server"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Osd is the object storage daemon.
type Osd struct {
	// Unique id of osd.
	id string

	cfg    *config.Osd
	server *server.Server
}

// New creates a osd object.
func New(cfg *config.Osd) (*Osd, error) {
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

	// Generate OSD ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating OSD UUID succeeded")

	s, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	o := &Osd{
		id:     cfg.ID,
		cfg:    cfg,
		server: s,
	}

	return o, nil
}

// Start starts the osd.
func (o *Osd) Start() {
	log.Info("Start OSD service ...")
	err := o.server.Start()
	if err != nil {
		log.Error(err)
	}
}

func (o *Osd) stop() {
	// Clean up osd works.
	log.Info("Stop OSD")
}
