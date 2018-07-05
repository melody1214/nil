package cluster

import (
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

func (s *service) updateClusterMap() error {
	txid, err := s.store.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}

	if err = s._updateClusterMap(txid); err != nil {
		s.store.Rollback(txid)
		return err
	}

	if err = s.store.Commit(txid); err != nil {
		s.store.Rollback(txid)
		return err
	}

	return nil
}

func (s *service) _updateClusterMap(txid repository.TxID) error {
	// Set new map version.
	ver, err := s.store.GetNewClusterMapVer(txid)
	if err != nil {
		return errors.Wrap(err, "failed to set new cmap map version")
	}

	// Create a cmap map with the new version.
	cm, err := s.createClusterMap(ver, txid)
	if err != nil {
		return errors.Wrap(err, "failed to create cmap map")
	}

	return s.cmapAPI.UpdateCMap(cm)
}

func (s *service) createClusterMap(ver cmap.Version, txid repository.TxID) (*cmap.CMap, error) {
	nodes, err := s.store.FindAllNodes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map nodes")
	}

	vols, err := s.store.FindAllVolumes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map volumes")
	}

	encGrps, err := s.store.FindAllEncGrps(txid)
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

func (s *service) updateDBByCMap(m *cmap.CMap) error {
	txid, err := s.store.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}

	if err = s.store.UpdateChangedCMap(txid, m); err != nil {
		s.store.Rollback(txid)
		return err
	}

	if err = s.store.Commit(txid); err != nil {
		s.store.Rollback(txid)
		return err
	}

	return nil
}
