package clustermap

import (
	"time"

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
	cm, err := h.createClusterMap(ver, txid)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster map")
	}

	return h.cMap.Update(cm)
}

func (h *handlers) createClusterMap(ver cmap.Version, txid repository.TxID) (*cmap.CMap, error) {
	nodes, err := h.store.FindAllNodes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster map nodes")
	}

	return &cmap.CMap{
		Version: ver,
		Time:    time.Now().UTC().String(),
		Nodes:   nodes,
	}, nil
}
