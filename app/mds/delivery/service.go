package delivery

import (
	"net"
	"net/rpc"
	"time"

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

	adh AdminHandlers
	auh AuthHandlers
	buh BucketHandlers
	clh ClustermapHandlers
	coh ConsensusHandlers
	meh MembershipHandlers
	reh RecoveryHandlers

	nilLayer        *nilmux.Layer
	raftLayer       *nilmux.Layer
	membershipLayer *nilmux.Layer

	nilMux    *nilmux.NilMux
	nilRPCSrv *rpc.Server
}

// NewDeliveryService creates a delivery service with necessary dependencies.
func NewDeliveryService(cfg *config.Mds, adh AdminHandlers, auh AuthHandlers, buh BucketHandlers, coh ConsensusHandlers, clh ClustermapHandlers, meh MembershipHandlers, reh RecoveryHandlers) (*Service, error) {
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

// AdminHandlers is the interface that provides admin domain's rpc handlers.
type AdminHandlers interface {
	Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error
	AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error
	GetLocalChain(req *nilrpc.GetLocalChainRequest, res *nilrpc.GetLocalChainResponse) error
	GetAllChain(req *nilrpc.GetAllChainRequest, res *nilrpc.GetAllChainResponse) error
	GetAllVolume(req *nilrpc.GetAllVolumeRequest, res *nilrpc.GetAllVolumeResponse) error
	RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error
}

// AuthHandlers is the interface that provides auth domain's rpc handlers.
type AuthHandlers interface {
	GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error
}

// BucketHandlers is the interface that provides bucket domain's rpc handlers.
type BucketHandlers interface {
	AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error
}

// ClustermapHandlers is the interface that provides clustermap domain's rpc handlers.
type ClustermapHandlers interface {
	GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error
	IsUpdated(req *nilrpc.ClusterMapIsUpdatedRequest, res *nilrpc.ClusterMapIsUpdatedResponse) error
}

// ConsensusHandlers is the interface that provides consensus domain's rpc handlers.
type ConsensusHandlers interface {
	Open(raftL *nilmux.Layer) error
	Stop() error
	Join() error
}

// MembershipHandlers is the interface that provides membership domain's rpc handlers.
type MembershipHandlers interface {
	GetMembershipList(req *nilrpc.GetMembershipListRequest, res *nilrpc.GetMembershipListResponse) error
	Create(swimL *nilmux.Layer) error
	Run()
}

// RecoveryHandlers is the interface that provides recovery domain's rpc handlers.
type RecoveryHandlers interface {
	Recover(req *nilrpc.RecoverRequest, res *nilrpc.RecoverResponse) error
	Rebalance(req *nilrpc.RebalanceRequest, res *nilrpc.RebalanceResponse) error
}
