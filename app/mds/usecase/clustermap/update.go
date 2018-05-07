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
		return errors.Wrap(err, "failed to set new cmap map version")
	}

	// Create a cmap map with the new version.
	cm, err := h.createClusterMap(ver, txid)
	if err != nil {
		return errors.Wrap(err, "failed to create cmap map")
	}

	return h.cmapAPI.UpdateCMap(cm)
}

func (h *handlers) createClusterMap(ver cmap.Version, txid repository.TxID) (*cmap.CMap, error) {
	nodes, err := h.store.FindAllNodes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map nodes")
	}

	vols, err := h.store.FindAllVolumes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map volumes")
	}

	encGrps, err := h.store.FindAllEncGrps(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map encoding groups")
	}

	return &cmap.CMap{
		Version: ver,
		Time:    cmap.Now(),
		Nodes:   nodes,
		Vols:    vols,
		EncGrps: encGrps,
	}, nil
}
