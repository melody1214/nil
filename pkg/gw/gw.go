package gw

import (
	"errors"

	"github.com/chanyoung/nil/pkg/gw/server"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Gw is the [project name] gateway server node.
type Gw struct {
	// Unique id of mds.
	id string

	// cfg is pointing the gateway config from command package.
	cfg *config.Gw

	// http server for client requests.
	server *server.Server
}

// New creates a gateway object.
func New(cfg *config.Gw) (*Gw, error) {
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

	// Generate gateway ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating gateway UUID succeeded")

	// Creates gateway server.
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	g := &Gw{
		id:     cfg.ID,
		cfg:    cfg,
		server: srv,
	}
	log.Info("Creating gateway object succeeded")

	return g, nil
}

// Start starts the gateway.
func (g *Gw) Start() {
	log.Info("Start gateway service ...")
	err := g.server.Start()
	if err != nil {
		log.Error(err)
	}
}

func (g *Gw) stop() {
	// Clean up gateway works.
	log.Info("Stop gateway")
}
