package cluster

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg     *config.Mds
	jFact   *jobFactory
	wPool   *workerPool
	store   Repository
	cmapAPI cmap.MasterAPI
}

// NewService creates a client service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.MasterAPI, s Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/cluster")

	service := &service{
		cfg:     cfg,
		jFact:   newJobFactory(newJobRepository(s)),
		wPool:   newWorkerPool(3, cmapAPI, newJobRepository(s)),
		store:   s,
		cmapAPI: cmapAPI,
	}
	service.runStateChangedObserver()

	return service
}

// Service is the interface that provides clustermap domain's service.
type Service interface {
	Join(raftL *nilmux.Layer) error
	Leave() error
	RPCHandler() RPCHandler
}

// Join joins the node to the global cluster.
func (s *service) Join(raftL *nilmux.Layer) error {
	if err := s.store.Open(raftL); err != nil {
		return err
	}

	// I'm the first node of this cluster, no need to join.
	if s.cfg.Raft.LocalClusterAddr == s.cfg.Raft.GlobalClusterAddr {
		return nil
	}

	return raftJoin(s.cfg.Raft.GlobalClusterAddr, s.cfg.Raft.LocalClusterAddr, s.cfg.Raft.LocalClusterRegion)
}

// Leave leaves the node from the global cluster.
func (s *service) Leave() error {
	return s.store.Close()
}

// RPCHandler returns the RPC handler which will handle
// the requests from the delivery layer.
func (s *service) RPCHandler() RPCHandler {
	// This is a trick to hide inadvertently exposed methods,
	// such as Join() or Leave().
	type handler struct{ RPCHandler }
	return handler{RPCHandler: s}
}

// RPCHandler is the interface that provides clustermap domain's rpc handlers.
type RPCHandler interface {
	GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error
	GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error
	// UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error
	RegisterVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error
	LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error
	GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error
	ListJob(req *nilrpc.MCLListJobRequest, res *nilrpc.MCLListJobResponse) error
}

// GetUpdateNoti returns when the cmap is updated or timeout.
func (s *service) GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error {
	notiC := s.cmapAPI.GetUpdatedNoti(cmap.Version(req.Version))

	timeout := time.After(10 * time.Minute)
	for {
		select {
		case <-notiC:
			return nil
		case <-timeout:
			return errors.New("timeout, try again")
		}
	}
}

// func (s *service) UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error {
// 	txid, err := s.store.Begin()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to start transaction")
// 	}

// 	if err = s.updateClusterMap(txid); err != nil {
// 		s.store.Rollback(txid)
// 		return err
// 	}
// 	if err = s.store.Commit(txid); err != nil {
// 		s.store.Rollback(txid)
// 		return err
// 	}

// 	s.rebalance()
// 	return nil
// }

// func (s *service) rebalance() error {
// 	conn, err := nilrpc.Dial(s.cfg.ServerAddr+":"+s.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	req := &nilrpc.MRERecoveryRequest{Type: nilrpc.Rebalance}
// 	res := &nilrpc.MRERecoveryResponse{}

// 	cli := rpc.NewClient(conn)
// 	return cli.Call(nilrpc.MdsRecoveryRecovery.String(), req, res)
// }

func calcMaxChain(volumeSize uint64) int {
	if volumeSize <= 0 {
		return 0
	}

	// Test, chain per 10MB,
	return int(volumeSize / 10)
}
