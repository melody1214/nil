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
	"github.com/chanyoung/nil/app/mds/usecase/membership"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
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

	adh admin.AdminHandlers
	auh auth.AuthHandlers
	buh bucket.BucketHandlers
	clh clustermap.ClustermapHandlers
	coh consensus.ConsensusHandlers
	meh membership.MembershipHandlers
	obh object.ObjectHandlers
	reh recovery.RecoveryHandlers

	nilLayer        *nilmux.Layer
	raftLayer       *nilmux.Layer
	membershipLayer *nilmux.Layer

	nilMux    *nilmux.NilMux
	nilRPCSrv *rpc.Server
}

// NewDeliveryService creates a delivery service with necessary dependencies.
func NewDeliveryService(cfg *config.Mds, adh admin.AdminHandlers, auh auth.AuthHandlers, buh bucket.BucketHandlers, coh consensus.ConsensusHandlers, clh clustermap.ClustermapHandlers, meh membership.MembershipHandlers, obh object.ObjectHandlers, reh recovery.RecoveryHandlers) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid argument")
	}
	logger = mlog.GetPackageLogger("app/mds/delivery")

	// Resolve gateway address.
	rAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, errors.Wrap(err, "resolve mds address failed")
	}

	// Create transport layers.
	nilL := nilmux.NewLayer(rpcTypeBytes(), rAddr, false)
	raftL := nilmux.NewLayer(raftTypeBytes(), rAddr, false)
	membershipL := nilmux.NewLayer(membershipTypeBytes(), rAddr, false)

	// Create a mux and register layers.
	m := nilmux.NewNilMux(cfg.ServerAddr+":"+cfg.ServerPort, &cfg.Security)
	m.RegisterLayer(nilL)
	m.RegisterLayer(raftL)
	m.RegisterLayer(membershipL)

	// Create swim server.
	if err := meh.Create(membershipL); err != nil {
		return nil, err
	}

	// Create rpc server.
	rpcSrv := rpc.NewServer()
	if err := rpcSrv.RegisterName(nilrpc.MdsAdminPrefix, adh); err != nil {
		return nil, err
	}
	if err := rpcSrv.RegisterName(nilrpc.MdsAuthPrefix, auh); err != nil {
		return nil, err
	}
	if err := rpcSrv.RegisterName(nilrpc.MdsBucketPrefix, buh); err != nil {
		return nil, err
	}
	if err := rpcSrv.RegisterName(nilrpc.MdsClustermapPrefix, clh); err != nil {
		return nil, err
	}
	if err := rpcSrv.RegisterName(nilrpc.MdsMembershipPrefix, meh); err != nil {
		return nil, err
	}
	if err := rpcSrv.RegisterName(nilrpc.MdsObjectPrefix, obh); err != nil {
		return nil, err
	}
	if err := rpcSrv.RegisterName(nilrpc.MdsRecoveryPrefix, reh); err != nil {
		return nil, err
	}

	return &Service{
		cfg: cfg,

		adh: adh,
		auh: auh,
		buh: buh,
		clh: clh,
		coh: coh,
		meh: meh,
		obh: obh,
		reh: reh,

		nilLayer:        nilL,
		raftLayer:       raftL,
		membershipLayer: membershipL,

		nilMux:    m,
		nilRPCSrv: rpcSrv,
	}, nil
}

func (s *Service) Run() error {
	go s.nilMux.ListenAndServeTLS()
	go s.serveNilRPC()

	if err := s.coh.Open(s.raftLayer); err != nil {
		return err
	}
	if err := s.coh.Join(); err != nil {
		return err
	}

	go s.meh.Run()
	go s.rebalancer()
	return nil
}

func (s *Service) Stop() error {
	return s.coh.Stop()
}

func (s *Service) rebalancer() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.rebalancer")

	// Make ticker for routinely rebalancing.
	t, err := time.ParseDuration(s.cfg.Rebalance)
	if err != nil {
		ctxLogger.Fatal(err)
	}
	rebalanceNoti := time.NewTicker(t)

	for {
		select {
		case <-rebalanceNoti.C:
			s.reh.Rebalance(
				&nilrpc.RebalanceRequest{},
				&nilrpc.RebalanceResponse{},
			)
		}
	}
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
