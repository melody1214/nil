package clustermap

import (
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

// TODO: with transaction.

func (h *handlers) updateClusterMap() error {
	// Set new map version.
	ver, err := h.store.GetNewClusterMapVer()
	if err != nil {
		return errors.Wrap(err, "failed to set new cluster map version")
	}

	// Create a cluster map with the new version.
	cm, err := h.createClusterMap(ver)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster map")
	}

	return h.cMap.Update(cmap.WithFile(cm))
}

func (h *handlers) createClusterMap(ver cmap.Version) (*cmap.CMap, error) {
	nodes, err := h.store.GetClusterMapNodes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster map nodes")
	}

	return &cmap.CMap{
		Version: ver,
		Nodes:   nodes,
	}, nil
}
