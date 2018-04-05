package clustermap

import (
	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type handlers struct {
	store Repository
	cMap  *cmap.CMap
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(s Repository) delivery.ClustermapHandlers {
	log = mlog.GetLogger().WithField("package", "mds/usecase/bucket")

	return &handlers{
		store: s,
		cMap:  cmap.New(),
	}
}

// GetClusterMap returns a current local cluster map.
func (h *handlers) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	cm, err := cmap.GetLatest()
	if err != nil {
		return err
	}

	res.Version = cm.Version
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

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := log.WithField("method", "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	h.cMap = m
}
