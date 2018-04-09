package clustermap

import (
	"errors"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store Repository
	cMap  *cmap.Controller
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cMap *cmap.Controller, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/clustermap")

	return &handlers{
		store: s,
		cMap:  cMap,
	}
}

// GetClusterMap returns a current local cluster map.
func (h *handlers) GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error {
	cm := h.cMap.LatestCMap()

	res.Version = cm.Version.Int64()
	for _, n := range cm.Nodes {
		res.Nodes = append(
			res.Nodes,
			nilrpc.ClusterNode{
				ID:   n.ID.Int64(),
				Name: n.Name,
				Addr: n.Addr,
				Type: n.Type.String(),
				Stat: n.Stat.String(),
			},
		)
	}

	return nil
}

// GetUpdateNoti returns when the cluster map is updated or timeout.
func (h *handlers) GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error {
	notiC := h.cMap.GetUpdatedNoti(cmap.Version(req.Version))

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

func (h *handlers) UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error {
	return h.updateClusterMap()
}

// Handlers is the interface that provides clustermap domain's rpc handlers.
type Handlers interface {
	GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error
	GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error
	UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error
}
