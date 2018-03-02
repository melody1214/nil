package server

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/swim"
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
	NilRPCHandler NilRPCHandler

	raftTransportLayer *nilmux.RaftTransportLayer
	raftLayer          *nilmux.Layer

	swimTransportLayer *nilmux.SwimTransportLayer
	swimLayer          *nilmux.Layer
	swimSrv            *swim.Server

	store *store.Store
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
	srv.raftTransportLayer = nilmux.NewRaftTransportLayer(srv.raftLayer)

	swimTypeBytes := []byte{
		0x03, // rpcSwim
	}
	srv.swimLayer = nilmux.NewLayer(swimTypeBytes, resolvedAddr, false)
	srv.swimTransportLayer = nilmux.NewSwimTransportLayer(srv.swimLayer)

	swimConf := swim.DefaultConfig()
	swimConf.ID = swim.ServerID(cfg.ID)
	swimConf.Address = swim.ServerAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	swimConf.Coordinator = swim.ServerAddress(cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(cfg.Swim.Period); err != nil {
		log.Error(err)
	} else {
		swimConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(cfg.Swim.Expire); err != nil {
		log.Error(err)
	} else {
		swimConf.PingExpire = t
	}
	swimConf.Type = swim.MDS

	srv.swimSrv, err = swim.NewServer(
		swimConf,
		srv.swimTransportLayer,
	)
	if err != nil {
		return nil, err
	}

	// Create a mux and register layers.
	srv.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	srv.nilMux.RegisterLayer(srv.nilLayer)
	srv.nilMux.RegisterLayer(srv.raftLayer)
	srv.nilMux.RegisterLayer(srv.swimLayer)

	// Create new raft store.
	srv.store = store.New(cfg, srv.raftTransportLayer)

	// Create nil RPC server.
	srv.nilRPCSrv = rpc.NewServer()
	if err := srv.registerNilRPCHandler(); err != nil {
		return nil, err
	}
	if err := srv.nilRPCSrv.RegisterName(nilrpc.MDSRPCPrefix, srv.NilRPCHandler); err != nil {
		return nil, err
	}

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

	// Starts swim service.
	sc := make(chan swim.PingError, 1)
	go s.swimSrv.Serve(sc)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case err := <-sc:
			log.WithFields(logrus.Fields{
				"server":       "swim",
				"message type": err.Type,
				"destID":       err.DestID,
			}).Error(err.Err)
		case <-sigc:
			log.Info("Received stop signal from OS")
			return s.stop()
		}
	}
}

// stop cleans up the services and shut down the server.
func (s *Server) stop() error {
	// Close store.
	s.store.Close()

	return nil
}

// join joins into the existing cluster, located at joinAddr.
// The joinAddr node must the leader state node.
func (s *Server) join(joinAddr, raftAddr, nodeID string) error {
	conn, err := nilrpc.Dial(joinAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.JoinRequest{
		RaftAddr: raftAddr,
		NodeID:   nodeID,
	}

	res := &nilrpc.JoinResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.Join.String(), req, res)
}

func (s *Server) registerNilRPCHandler() (err error) {
	s.NilRPCHandler, err = newNilRPCHandler(s)
	return
}
