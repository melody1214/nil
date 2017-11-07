package mds

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/db"
	"github.com/chanyoung/nil/pkg/mds/server"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Mds is the [project name] meta-data server node.
type Mds struct {
	// Unique id of mds.
	id string

	// cfg is pointing the mds config from command package.
	cfg *config.Mds

	// RPC server.
	server *server.Server

	// Key/Value store for metadata.
	db db.DB
}

// New creates a mds object.
func New(cfg *config.Mds) (*Mds, error) {
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

	// Generate MDS ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating MDS UUID succeeded")

	m := &Mds{
		id:     cfg.ID,
		cfg:    cfg,
		server: server.New(cfg),
		db:     db.New(),
	}
	log.Info("Creating MDS object succeeded")

	return m, nil
}

// Start starts the mds.
func (m *Mds) Start() {
	// Make channel for Ctrl-C or other terminate signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Make channel for message from server.
	mc := make(chan error, 1)
	go m.server.Start(mc)

	// Join membership list.
	if err := m.server.JoinMembership(); err != nil {
		log.Fatal(err)
	}

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
