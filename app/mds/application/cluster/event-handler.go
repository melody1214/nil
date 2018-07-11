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

	j, err := s.jFact.create(e, req.Volumes)
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
	lastCMap := s.cmapAPI.GetLatestCMap()

	for {
		notiC := s.cmapAPI.GetStateChangedNoti()
		select {
		case <-notiC:
			changedCMap := s.cmapAPI.GetLatestCMap()

			// Update volume max eg.
			for i, v := range changedCMap.Vols {
				changedCMap.Vols[i].MaxEG = calcMaxEG(v.Size)
			}
			if err := s.updateDBByCMap(&changedCMap); err != nil {
				fmt.Printf("\n%+v\n", err)
			}

			events := extractEventsFromCMap(&lastCMap, &changedCMap)
			for _, e := range events {
				j, err := s.jFact.create(e, nil)
				if err != nil {
					// TODO: handling errors.
					fmt.Printf("\n%+v\n", err)
					continue
				}

				if j.Type == Iterative {
					s.wPool.dispatchNow(j)
				}
			}

			lastCMap = changedCMap
		}
	}
}
