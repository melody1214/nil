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

// ListJob returns a current job list.
func (s *service) ListJob(req *nilrpc.MCLListJobRequest, res *nilrpc.MCLListJobResponse) error {
	res.List = s.store.ListJob()
	return nil
}

// RegisterVolume receives a new volume information from ds and register it to the database.
func (s *service) RegisterVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	e := newEvent(RegisterVolume, NoAffectedEG)

	j, err := s.jFact.create(e, &req.Volume)
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
		res.ID = req.Volume.ID
		return err
	case <-timeout:
		// TODO: j.stop()
		return fmt.Errorf("timeout failed")
	}
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

// LocalJoin handles the join request from the same local cluster nodes.
func (s *service) LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error {
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
}

// GlobalJoin handles the join request from the other raft nodes.
func (s *service) GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}
	return s.store.GlobalJoin(req.RaftAddr, req.NodeID)
}

func (s *service) runStateChangedObserver() {
	go func() {
		for {
			notiC := s.cmapAPI.GetStateChangedNoti()
			select {
			case <-notiC:
				// Handle state changed.
			}
		}
	}()
}
