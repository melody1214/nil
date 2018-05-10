package cluster

import (
	"fmt"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetClusterMap returns a current local cmap.
func (s *service) GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error {
	res.ClusterMap = s.cmapAPI.GetLatestCMap()
	return nil
}

// RegisterVolume receives a new volume information from ds and register it to the database.
func (s *service) RegisterVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	// if req.ID != "" {
	// 	return fmt.Errorf("not allowed to update volume by rpc")
	// }

	// if err := s.insertNewVolume(req, res); err != nil {
	// 	return err
	// }

	// go s.jFact.create(newEvent(Rebalance, 0))
	// return nil
	// // If the id field of request is empty, then the ds
	// // tries to get an id of volume.
	// if req.ID == "" {
	// 	return s.insertNewVolume(req, res)
	// }
	// return s.updateVolume(req, res)
	return nil
}

func (s *service) updateVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	// ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateVolume")

	// q := fmt.Sprintf(
	// 	`
	// 	UPDATE volume
	// 	SET vl_status='%s', vl_size='%d', vl_free='%d', vl_used='%d', vl_max_encoding_group='%d', vl_speed='%s'
	// 	WHERE vl_id in ('%s')
	// 	`, req.Status, req.Size, req.Free, req.Used, calcMaxChain(req.Size), req.Speed, req.ID,
	// )

	// _, err := s.store.Execute(repository.NotTx, q)
	// if err != nil {
	// 	ctxLogger.Error(err)
	// 	return err
	// }

	// return s.UpdateClusterMap(nil, nil)
	return nil
}

func (s *service) insertNewVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	// ctxLogger := mlog.GetMethodLogger(logger, "handlers.insertNewVolume")

	// q := fmt.Sprintf(
	// 	`
	// 	INSERT INTO volume (vl_node, vl_status, vl_size, vl_free, vl_used, vl_encoding_group, vl_max_encoding_group, vl_speed)
	// 	SELECT node_id, '%s', '%d', '%d', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
	// 	`, req.Status, req.Size, req.Free, req.Used, 0, calcMaxChain(req.Size), req.Speed, req.Ds,
	// )

	// r, err := s.store.Execute(repository.NotTx, q)
	// if err != nil {
	// 	ctxLogger.Error(err)
	// 	return err
	// }

	// id, err := r.LastInsertId()
	// if err != nil {
	// 	ctxLogger.Error(err)
	// 	return err
	// }
	// res.ID = strconv.FormatInt(id, 10)

	return nil
}

// LocalJoin handles the join request from the same local cluster nodes.
func (s *service) LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error {
	// Make event structure.
	e := newEvent(LocalJoin, NoAffectedEG)

	j, err := s.jFact.create(e, req.Node)
	if err != nil {
		return err
	}

	s.wPool.dispatchNow(j)

	waitC, err := j.getWaitChannel()
	if err != nil {
		return err
	}

	timeout := time.After(2 * time.Minute)
	select {
	case err = <-waitC:
		return err
	case <-timeout:
		// TODO: j.stop()
		return fmt.Errorf("timeout failed")
	}

	// Add new node into the database.
	// if err := s.store.LocalJoin(req.Node); err != nil {
	// 	return errors.Wrap(err, "failed to add new node into the database")
	// }

	// Just update the cluster map.
	// Removed the rebalance call because the only case of local join is
	// the ds without any volumes.
	// s.updateClusterMap()
}

// GlobalJoin handles the join request from the other raft nodes.
func (s *service) GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}
	return s.store.GlobalJoin(req.RaftAddr, req.NodeID)
}
