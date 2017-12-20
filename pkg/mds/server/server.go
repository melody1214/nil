package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/mds/mysql"
	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server serve RPCs.
type Server struct {
	cfg *config.Mds

	raftTr *RaftTransportLayer

	httpTr  *HTTPTransportLayer
	httpSrv *http.Server

	store *store.Store
	db    *mysql.MySQL
}

// New creates a rpc server object.
func New(cfg *config.Mds) (*Server, error) {
	log = mlog.GetLogger()

	// Create transport layers.
	localAddr, err := net.ResolveTCPAddr("tcp", cfg.Raft.LocalClusterAddr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:    cfg,
		raftTr: newRaftTransportLayer(localAddr),
		httpTr: newHTTPTransportLayer(localAddr),
	}

	s.httpSrv = &http.Server{
		Handler:        s,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Create new raft store.
	s.store = store.New(&cfg.Raft, &cfg.Security, s.raftTr)

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
	// Start tcp listen and serve.
	go s.listenAndServeTLS()
	go s.httpSrv.Serve(s.httpTr)

	// Start raft service.
	if err := s.store.Open(); err != nil {
		return err
	}
	log.Info("raft started successfully")

	if s.cfg.Raft.LocalClusterAddr != s.cfg.Raft.GlobalClusterAddr {
		// Join to the existing raft cluster.
		if err := s.join(s.cfg.Raft.GlobalClusterAddr,
			s.cfg.Raft.LocalClusterAddr,
			s.cfg.Raft.LocalClusterRegion); err != nil {
			return errors.Wrap(err, "open raft: failed to join existing cluster")
		}
	}

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

// join joins into the existing cluster, located at joinAddr.
// The joinAddr node must the leader state node.
func (s *Server) join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}

	hc := s.httpTr.GetHTTPClient()
	resp, err := hc.Post(
		fmt.Sprintf("http://%s/join", joinAddr),
		"application/raft",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
