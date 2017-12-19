package server

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/mds/mysql"
	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server serve RPCs.
type Server struct {
	cfg   *config.Mds
	rtl   *RaftTransportLayer
	store *store.Store
	db    *mysql.MySQL
}

// New creates a rpc server object.
func New(cfg *config.Mds) (*Server, error) {
	log = mlog.GetLogger()

	s := &Server{
		cfg: cfg,
	}

	// Create new raft store.
	raftAddr, err := net.ResolveTCPAddr("tcp", cfg.Raft.LocalClusterAddr)
	if err != nil {
		return nil, err
	}
	s.rtl = newRaftTransportLayer(raftAddr)
	s.store = store.New(&cfg.Raft, &cfg.Security, s.rtl)

	// Connect and initiate to mysql server.
	db, err := mysql.New(s.cfg)
	if err != nil {
		return nil, err
	}
	s.db = db

	return s, nil
}

// Start starts to listen and serve RPCs.
func (s *Server) Start() error {
	go s.listenAndServeTLS()

	if err := s.store.Open(); err != nil {
		return err
	}
	log.Info("raft started successfully")

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			log.Info("Received stop signal from OS")
			return s.stop()
		}
	}
}

// stop cleans up the services and shut down the server.
func (s *Server) stop() error {
	// Close mysql connection.
	s.db.Close()

	return nil
}
