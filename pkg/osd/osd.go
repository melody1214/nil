package osd

import (
	"errors"

	"github.com/chanyoung/nil/pkg/rest"
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

	cfg *config.Osd

	restServer *rest.Server
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

	o := &Osd{
		id:         cfg.ID,
		cfg:        cfg,
		restServer: rest.NewServer(),
	}

	return o, nil
}

// Start starts the osd.
func (o *Osd) Start() {
	// if err := o.restServer.Start(); err != nil {
	// 	log.Fatal(err)
	// }

	// // Wait until Ctrl-C or other terminate signal is received.
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	// <-sc

	// o.stop()
}

func (o *Osd) stop() {
	// Clean up osd works.
	// log.Println("stop osd")
}
