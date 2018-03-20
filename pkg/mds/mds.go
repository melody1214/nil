package mds

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/mds/rpchandling"
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
	NilRPCHandler rpchandling.NilRPCHandler

	raftTransportLayer *nilmux.RaftTransportLayer
	raftLayer          *nilmux.Layer

	swimTransportLayer *nilmux.SwimTransportLayer
	swimLayer          *nilmux.Layer
	swimSrv            *swim.Server

	store *store.Store
}

// New creates a mds object.
func New(cfg *config.Mds) (*Mds, error) {
	// 1. Setting logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return nil, err
	}
	log = mlog.GetLogger()
	if log == nil {
		return nil, errors.New("nil logger object")
	}
	log.WithField("location", cfg.LogLocation).Info("Setting logger succeeded")

	// 2. Generate MDS ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating MDS UUID succeeded")

	// 3. Creates metadata server object with the given config.
	m := &Mds{cfg: cfg}

	// 4. Resolve mds address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	// 5. Create a rpc layer.
	m.nilLayer = nilmux.NewLayer(rpchandling.TypeBytes(), resolvedAddr, false)

	// 6. Create a raft layer.
	raftTypeBytes := []byte{
		0x01, // rpcRaft
	}
	m.raftLayer = nilmux.NewLayer(raftTypeBytes, resolvedAddr, false)
	m.raftTransportLayer = nilmux.NewRaftTransportLayer(m.raftLayer)

	// 7. Create a swim layer.
	swimTypeBytes := []byte{
		0x03, // rpcSwim
	}
	m.swimLayer = nilmux.NewLayer(swimTypeBytes, resolvedAddr, false)
	m.swimTransportLayer = nilmux.NewSwimTransportLayer(m.swimLayer)

	// 8. Load a swim config.
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

	// 9. Create a swim server.
	m.swimSrv, err = swim.NewServer(
		swimConf,
		m.swimTransportLayer,
	)
	if err != nil {
		return nil, err
	}

	// 10. Create a mux and register layers.
	m.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	m.nilMux.RegisterLayer(m.nilLayer)
	m.nilMux.RegisterLayer(m.raftLayer)
	m.nilMux.RegisterLayer(m.swimLayer)

	// 11. Create new raft store.
	m.store = store.New(cfg, m.raftTransportLayer)

	// 12. Create nil RPC server.
	m.nilRPCSrv = rpc.NewServer()
	m.NilRPCHandler, err = rpchandling.New(m.store)
	if err != nil {
		return nil, err
	}
	if err := m.nilRPCSrv.RegisterName(nilrpc.MDSRPCPrefix, m.NilRPCHandler); err != nil {
		return nil, err
	}

	// 13. Prepare the initial cluster map.
	if err := cmap.Initial(cfg.ServerAddr + ":" + cfg.ServerPort); err != nil {
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

func (m *Mds) serveNilRPC(l *nilmux.Layer) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go m.nilRPCSrv.ServeConn(conn)
	}
}
