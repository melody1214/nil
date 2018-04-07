package clustermap

import (
	"github.com/chanyoung/nil/app/mds/delivery"
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
func NewHandlers(cMap *cmap.Controller, s Repository) delivery.ClustermapHandlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/clustermap")

	return &handlers{
		store: s,
		cMap:  cMap,
	}
}

// GetClusterMap returns a current local cluster map.
func (h *handlers) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
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
