package mds

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
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Mds is the [project name] meta-data server node.
type Mds struct {
	// cfg is pointing the mds config from command package.
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

// New creates a mds object.
func New(cfg *config.Mds) (*Mds, error) {
	// Setting logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return nil, err
	}
	log = mlog.GetLogger()
	if log == nil {
		return nil, errors.New("nil logger object")
	}
	log.WithField("location", cfg.LogLocation).Info("Setting logger succeeded")

	// Generate MDS ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating MDS UUID succeeded")

	m := &Mds{cfg: cfg}

	// Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	// Create a rpc layer.
	rpcTypeBytes := []byte{
		0x02, // rpcNil
	}
	m.nilLayer = nilmux.NewLayer(rpcTypeBytes, resolvedAddr, false)

	// Create a raft layer.
	raftTypeBytes := []byte{
		0x01, // rpcRaft
	}
	m.raftLayer = nilmux.NewLayer(raftTypeBytes, resolvedAddr, false)
	m.raftTransportLayer = nilmux.NewRaftTransportLayer(m.raftLayer)

	swimTypeBytes := []byte{
		0x03, // rpcSwim
	}
	m.swimLayer = nilmux.NewLayer(swimTypeBytes, resolvedAddr, false)
	m.swimTransportLayer = nilmux.NewSwimTransportLayer(m.swimLayer)

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

	m.swimSrv, err = swim.NewServer(
		swimConf,
		m.swimTransportLayer,
	)
	if err != nil {
		return nil, err
	}

	// Create a mux and register layers.
	m.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	m.nilMux.RegisterLayer(m.nilLayer)
	m.nilMux.RegisterLayer(m.raftLayer)
	m.nilMux.RegisterLayer(m.swimLayer)

	// Create new raft store.
	m.store = store.New(cfg, m.raftTransportLayer)

	// Create nil RPC server.
	m.nilRPCSrv = rpc.NewServer()
	if err := m.registerNilRPCHandler(); err != nil {
		return nil, err
	}
	if err := m.nilRPCSrv.RegisterName(nilrpc.MDSRPCPrefix, m.NilRPCHandler); err != nil {
		return nil, err
	}

	log.Info("Creating MDS object succeeded")

	return m, nil
}

// Start starts the mds.
func (m *Mds) Start() {
	log.Info("Start MDS service ...")

	// Start tcp listen and serve.
	go m.nilMux.ListenAndServeTLS()
	go m.serveNilRPC(m.nilLayer)

	// Start raft service.
	if err := m.store.Open(); err != nil {
		log.Fatal(err)
	}
	log.Info("raft started successfully")

	if m.cfg.Raft.LocalClusterAddr != m.cfg.Raft.GlobalClusterAddr {
		// Join to the existing raft cluster.
		if err := m.join(m.cfg.Raft.GlobalClusterAddr,
			m.cfg.Raft.LocalClusterAddr,
			m.cfg.Raft.LocalClusterRegion); err != nil {
			log.Fatal(errors.Wrap(err, "open raft: failed to join existing cluster"))
		}
	}

	// Starts swim service.
	sc := make(chan swim.PingError, 1)
	go m.swimSrv.Serve(sc)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case err := <-sc:
			m.recover(err)
		case <-sigc:
			log.Info("Received stop signal from OS")
			m.stop()
			return
		}
	}
}

// stop cleans up the services and shut down the server.
func (m *Mds) stop() error {
	log.Info("Stop MDS")

	// Clean up mds works.
	// Close store.
	m.store.Close()

	return nil
}

// join joins into the existing cluster, located at joinAddr.
// The joinAddr node must the leader state node.
func (m *Mds) join(joinAddr, raftAddr, nodeID string) error {
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

func (m *Mds) registerNilRPCHandler() (err error) {
	m.NilRPCHandler, err = newNilRPCHandler(m)
	return
}
