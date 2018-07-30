package cluster

import (
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

var region string

type service struct {
	cfg     *config.Mds
	store   Repository
	cmapAPI cmap.MasterAPI
}

// NewService creates a client service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.MasterAPI, s Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/cluster")

	region = cfg.Raft.LocalClusterRegion
	service := &service{
		cfg:     cfg,
		store:   s,
		cmapAPI: cmapAPI,
	}

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
	// if err := s.store.Open(raftL); err != nil {
	// 	return err
	// }

	// // I'm the first node of this cluster, no need to join.
	// if s.cfg.Raft.LocalClusterAddr == s.cfg.Raft.GlobalClusterAddr {
	// 	return nil
	// }

	// return raftJoin(s.cfg.Raft.GlobalClusterAddr, s.cfg.Raft.LocalClusterAddr, s.cfg.Raft.LocalClusterRegion)
	return nil
}

// Leave leaves the node from the global cluster.
func (s *service) Leave() error {
	// return s.store.Close()
	return nil
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
	LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error
	GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error
}
