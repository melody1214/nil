package mds

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/db"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
)

var log *mlog.Log

// Mds is the [project name] meta-data server node.
type Mds struct {
	// Unique id of mds.
	id string

	// cfg is pointing the mds config from command package.
	cfg *Config

	// RPC server.
	server *server

	// Key/Value store for metadata.
	db db.DB
}

// New creates a mds object.
func New(cfg *Config) (*Mds, error) {
	// Setting logger.
	l, err := mlog.New(cfg.LogLocation)
	if err != nil {
		return nil, err
	}

	log = l
	log.WithField("location", cfg.LogLocation).Info("Setting log lcation")

	// Generate MDS ID.
	cfg.ID = uuid.Gen()

	m := &Mds{
		id:     cfg.ID,
		cfg:    cfg,
		server: newServer(cfg),
		db:     db.New(),
	}

	return m, nil
}

// Start starts the mds.
func (m *Mds) Start() {
	// Make channel for Ctrl-C or other terminate signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Make channel for message from server.
	mc := make(chan error, 1)
	go m.server.start(mc)

	log.Info("MDS running ...")
	for {
		select {
		case <-sc:
			log.Info("Received stop signal from OS")
			m.stop()
			return
		case err := <-mc:
			log.Error(err)
		}
	}
}

func (m *Mds) stop() {
	// Clean up mds works.
	log.Info("Stop MDS")
}
