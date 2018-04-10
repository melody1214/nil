package clustermap

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

func (h *handlers) updateClusterMap(txid repository.TxID) error {
	// Set new map version.
	ver, err := h.store.GetNewClusterMapVer(txid)
	if err != nil {
		return errors.Wrap(err, "failed to set new cluster map version")
	}

	// Create a cluster map with the new version.
	cm, err := h.createClusterMap(ver)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster map")
	}

	return h.cMap.Update(cm)
}

func (h *handlers) createClusterMap(ver cmap.Version) (*cmap.CMap, error) {
	nodes, err := h.store.FindAllNodes(repository.NotTx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster map nodes")
	}

	return &cmap.CMap{
		Version: ver,
		Nodes:   nodes,
	}, nil
}
