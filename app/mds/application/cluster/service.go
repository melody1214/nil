package cluster

import (
	"github.com/chanyoung/nil/app/mds/domain/model/region"
	"github.com/chanyoung/nil/app/mds/domain/service/raft"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg     *config.Mds
	rs      raft.Service
	rr      region.Repository
	cmapAPI cmap.MasterAPI
}

// NewService creates a client service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.MasterAPI, rs raft.Service, rr region.Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/cluster")

	service := &service{
		cfg:     cfg,
		rs:      rs,
		rr:      rr,
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
