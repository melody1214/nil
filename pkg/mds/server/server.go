package server

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/mds/mysql"
	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server serve RPCs.
type Server struct {
	cfg *config.Mds

	nilMux        *nilmux.NilMux
	nilLayer      *nilmux.Layer
	nilRPCSrv     *rpc.Server
	NilRPCHandler *NilRPCHandler

	raftTransportLayer *raftTransportLayer
	raftLayer          *nilmux.Layer

	store *store.Store
	db    *mysql.MySQL
}

// New creates a rpc server object.
func New(cfg *config.Mds) (*Server, error) {
	log = mlog.GetLogger()

	// Resolve gateway addres.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		cfg: cfg,
	}

	// Create a rpc layer.
	rpcTypeBytes := []byte{
		0x02, // rpcNil
	}
	srv.nilLayer = nilmux.NewLayer(rpcTypeBytes, resolvedAddr, false)

	// Create a raft layer.
	raftTypeBytes := []byte{
		0x01, // rpcRaft
	}
	srv.raftLayer = nilmux.NewLayer(raftTypeBytes, resolvedAddr, false)
	srv.raftTransportLayer = newRaftTransportLayer(srv.raftLayer)

	// Create a mux and register layers.
	srv.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	srv.nilMux.RegisterLayer(srv.nilLayer)
	srv.nilMux.RegisterLayer(srv.raftLayer)

	// Create new raft store.
	// srv.store = store.New(&cfg.Raft, &cfg.Security, srv.raftTr)
	srv.store = store.New(&cfg.Raft, &cfg.Security, srv.raftTransportLayer)

	// // Create nil RPC server.
	// srv.nilRPC = newNilRPC(srv)
	srv.nilRPCSrv = rpc.NewServer()
	srv.NilRPCHandler = newNilRPCHandler(srv)
	if err := srv.nilRPCSrv.Register(srv.NilRPCHandler); err != nil {
		return nil, err
	}

	// Connect and initiate to mysql server.
	db, err := mysql.New(srv.cfg)
	if err != nil {
		return nil, err
	}
	srv.db = db

	return srv, nil
}

// Start starts to listen and serve RPCs.
func (s *Server) Start() error {
	// Start tcp listen and serve.
	go s.nilMux.ListenAndServeTLS()
	go s.serveNilRPC(s.nilLayer)

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
	conn, err := dialNilRPC(joinAddr, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &JoinRequest{
		RaftAddr: raftAddr,
		NodeID:   nodeID,
	}

	res := &JoinResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call("NilRPCHandler.HandleJoin", req, res)
}
