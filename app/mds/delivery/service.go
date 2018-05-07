package delivery

import (
	"net"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/mds/usecase/admin"
	"github.com/chanyoung/nil/app/mds/usecase/auth"
	"github.com/chanyoung/nil/app/mds/usecase/bucket"
	"github.com/chanyoung/nil/app/mds/usecase/clustermap"
	"github.com/chanyoung/nil/app/mds/usecase/consensus"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/chanyoung/nil/pkg/cluster"
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

	adh admin.Handlers
	auh auth.Handlers
	buh bucket.Handlers
	clh clustermap.Handlers
	coh consensus.Handlers
	cls *cluster.Service
	obh object.Handlers
	reh recovery.Handlers

	nilLayer        *nilmux.Layer
	raftLayer       *nilmux.Layer
	membershipLayer *nilmux.Layer

	nilMux    *nilmux.NilMux
	nilRPCSrv *rpc.Server
}

// SetupDeliveryService bootstraps a delivery service with necessary dependencies.
func SetupDeliveryService(cfg *config.Mds, adh admin.Handlers, auh auth.Handlers, buh bucket.Handlers, coh consensus.Handlers, clh clustermap.Handlers, cls *cluster.Service, obh object.Handlers, reh recovery.Handlers) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid argument")
	}
	logger = mlog.GetPackageLogger("app/mds/delivery")

	s := &Service{
		cfg: cfg,

		adh: adh,
		auh: auh,
		buh: buh,
		clh: clh,
		coh: coh,
		cls: cls,
		obh: obh,
		reh: reh,
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

	// // Create swim server.
	// if err := meh.Create(membershipL); err != nil {
	// 	return nil, err
	// }

	// Create rpc server.
	s.nilRPCSrv = rpc.NewServer()
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsAdminPrefix, s.adh); err != nil {
		return nil, err
	}
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsAuthPrefix, s.auh); err != nil {
		return nil, err
	}
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsBucketPrefix, s.buh); err != nil {
		return nil, err
	}
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsClustermapPrefix, s.clh); err != nil {
		return nil, err
	}
	// if err := rpcSrv.RegisterName(nilrpc.MdsMembershipPrefix, meh); err != nil {
	// 	return nil, err
	// }
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsObjectPrefix, s.obh); err != nil {
		return nil, err
	}
	if err := s.nilRPCSrv.RegisterName(nilrpc.MdsRecoveryPrefix, s.reh); err != nil {
		return nil, err
	}

	// Run the delivery server.
	if err := s.run(); err != nil {
		return nil, err
	}

	// Setup the membership server and run.
	clusterConf := cluster.DefaultConfig()
	clusterConf.Name = cluster.NodeName(cfg.ID)
	clusterConf.Address = cluster.NodeAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	clusterConf.Coordinator = cluster.NodeAddress(cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(cfg.Swim.Period); err == nil {
		clusterConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(cfg.Swim.Expire); err == nil {
		clusterConf.PingExpire = t
	}
	clusterConf.Type = cluster.MDS
	if err := s.cls.StartMembershipServer(*clusterConf, nilmux.NewSwimTransportLayer(s.membershipLayer)); err != nil {
		return nil, err
	}
	// Join the local cluster.
	conn, err := nilrpc.Dial(clusterConf.Coordinator.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &nilrpc.MCLJoinRequest{
		Node: cluster.Node{
			Name: clusterConf.Name,
			Type: clusterConf.Type,
			Stat: cluster.Alive,
			Addr: clusterConf.Address,
		},
	}
	res := &nilrpc.MCLJoinResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClustermapJoin.String(), req, res); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) run() error {
	go s.nilMux.ListenAndServeTLS()
	go s.serveNilRPC()

	if err := s.coh.Open(s.raftLayer); err != nil {
		return err
	}
	return s.coh.Join()
	// if err := s.coh.Join(); err != nil {
	// return err
	// }
	// return s.meh.Run(s.membershipLayer)
}

func (s *Service) Stop() error {
	return s.coh.Stop()
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
