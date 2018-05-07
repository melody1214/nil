package clustermap

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cluster"
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

	return h.clusterAPI.UpdateCMap(cm)
}

func (h *handlers) createClusterMap(ver cluster.CMapVersion, txid repository.TxID) (*cluster.CMap, error) {
	nodes, err := h.store.FindAllNodes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster map nodes")
	}

	vols, err := h.store.FindAllVolumes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster map volumes")
	}

	encGrps, err := h.store.FindAllEncGrps(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster map encoding groups")
	}

	return &cluster.CMap{
		Version: ver,
		Time:    cluster.CMapNow(),
		Nodes:   nodes,
		Vols:    vols,
		EncGrps: encGrps,
	}, nil
}
