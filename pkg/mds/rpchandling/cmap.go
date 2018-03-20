package rpchandling

import (
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetClusterMap returns a current local cluster map.
func (h *Handler) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
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
