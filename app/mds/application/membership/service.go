package membership

import (
	"github.com/chanyoung/nil/app/mds/domain/model/clustermap"
	"github.com/chanyoung/nil/app/mds/domain/model/region"
	"github.com/chanyoung/nil/app/mds/domain/service/raft"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Entry
	opened = false
)

type service struct {
	cfg     *config.Mds
	rs      raft.Service
	rr      region.Repository
	cr      clustermap.Repository
	cmapAPI cmap.MasterAPI
}

// NewService creates a client service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.MasterAPI, rs raft.Service, rr region.Repository, cr clustermap.Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/cluster")

	s := &service{
		cfg:     cfg,
		rs:      rs,
		rr:      rr,
		cr:      cr,
		cmapAPI: cmapAPI,
	}

	return s
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
	GetClusterMap(req *nilrpc.MMEGetClusterMapRequest, res *nilrpc.MMEGetClusterMapResponse) error
	GetUpdateNoti(req *nilrpc.MMEGetUpdateNotiRequest, res *nilrpc.MMEGetUpdateNotiResponse) error
	LocalJoin(req *nilrpc.MMELocalJoinRequest, res *nilrpc.MMELocalJoinResponse) error
	GlobalJoin(req *nilrpc.MMEGlobalJoinRequest, res *nilrpc.MMEGlobalJoinResponse) error
	UpdateNode(req *nilrpc.MMEUpdateNodeRequest, res *nilrpc.MMEUpdateNodeResponse) error
}
