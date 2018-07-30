package delivery

import (
	"net"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/mds/application/cluster"
	"github.com/chanyoung/nil/app/mds/application/gencoding"
	"github.com/chanyoung/nil/app/mds/application/object"
	"github.com/chanyoung/nil/app/mds/application/user"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type Service struct {
	cfg *config.Mds

	uss user.Service
	cls cluster.Service
	cms *cmap.Service
	obh object.Handlers
	ges gencoding.Service

	nilLayer        *nilmux.Layer
	raftLayer       *nilmux.Layer
	membershipLayer *nilmux.Layer

	nilMux    *nilmux.NilMux
	nilRPCSrv *rpc.Server
}

// SetupDeliveryService bootstraps a delivery service with necessary dependencies.
func SetupDeliveryService(cfg *config.Mds, uss user.Service, cls cluster.Service, cms *cmap.Service, obh object.Handlers, ges gencoding.Service) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid argument")
	}
	logger = mlog.GetPackageLogger("app/mds/delivery")

	s := &Service{
		cfg: cfg,

		uss: uss,
		cls: cls,
		cms: cms,
		obh: obh,
		ges: ges,
	}

	// Resolve gateway address.
	rAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, errors.Wrap(err, "resolve mds address failed")
	}

	// Create transport layers.
	s.nilLayer = nilmux.NewLayer(rpcTypeBytes(), rAddr, false)
	s.raftLayer = nilmux.NewLayer(raftTypeBytes(), rAddr, false)
	s.membershipLayer = nilmux.NewLayer(membershipTypeBytes(), rAddr, false)

	// Create a mux and register layers.
	s.nilMux = nilmux.NewNilMux(cfg.ServerAddr+":"+cfg.ServerPort, &cfg.Security)
	s.nilMux.RegisterLayer(s.nilLayer)
	s.nilMux.RegisterLayer(s.raftLayer)
	s.nilMux.RegisterLayer(s.membershipLayer)

	// Create rpc server.
	s.nilRPCSrv = rpc.NewServer()
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsUserPrefix, s.uss); err != nil {
		return nil, err
	}
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsClusterPrefix, s.cls.RPCHandler()); err != nil {
		return nil, err
	}
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsObjectPrefix, s.obh); err != nil {
		return nil, err
	}
	// if err := s.nilRPCSrv.RegisterName(nilrpc.MdsGencodingPrefix, s.ges); err != nil {
	// 	return nil, err
	// }

	// Setup the membership server and run.
	cmapConf := cmap.DefaultConfig()
	cmapConf.Name = cmap.NodeName(cfg.ID)
	cmapConf.Address = cmap.NodeAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	cmapConf.Coordinator = cmap.NodeAddress(cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(cfg.Swim.Period); err == nil {
		cmapConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(cfg.Swim.Expire); err == nil {
		cmapConf.PingExpire = t
	}
	cmapConf.Type = cmap.MDS

	// Setup tls listener.
	if err := s.nilMux.ListenAndServeTLS(); err != nil {
		return nil, err
	}

	// Setup database and join to global.
	if err := s.cls.Join(s.raftLayer); err != nil {
		return nil, err
	}

	// Join to the local cmap.
	if err := s.joinToLocal(cmapConf); err != nil {
		return nil, err
	}

	// Setup rpc server.
	go s.serveNilRPC()

	// Start local membership protocol.
	if err := s.cms.StartMembershipServer(*cmapConf, nilmux.NewSwimTransportLayer(s.membershipLayer)); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) joinToLocal(cmapConf *cmap.Config) error {
	// Join the local cmap.
	req := &nilrpc.MCLLocalJoinRequest{
		Node: cmap.Node{
			Name: cmapConf.Name,
			Type: cmapConf.Type,
			Stat: cmap.NodeAlive,
			Addr: cmapConf.Address,
		},
	}
	res := &nilrpc.MCLLocalJoinResponse{}

	// I am the very firstman of the land.
	if cmapConf.Address.String() == s.cfg.ServerAddr+":"+s.cfg.ServerPort {
		return s.cls.RPCHandler().LocalJoin(req, res)
	}

	// Ask to join other node.
	conn, err := nilrpc.Dial(cmapConf.Coordinator.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	cli := rpc.NewClient(conn)
	defer cli.Close()
	return cli.Call(nilrpc.MdsClusterLocalJoin.String(), req, res)
}

func (s *Service) Stop() error {
	return s.cls.Leave()
}

func (s *Service) serveNilRPC() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.serveNilRPC")

	for {
		conn, err := s.nilLayer.Accept()
		if err != nil {
			ctxLogger.Error(err)
			return
		}
		go s.nilRPCSrv.ServeConn(conn)
	}
}
