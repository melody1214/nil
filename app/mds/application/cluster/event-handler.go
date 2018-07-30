package cluster

import (
	"errors"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetClusterMap returns a current local cmap.
func (s *service) GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error {
	res.ClusterMap = s.cmapAPI.GetLatestCMap()
	return nil
}

// LocalJoin handles the join request from the same local cluster nodes.
func (s *service) LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error {
	return nil
}

// GlobalJoin handles the join request from the other raft nodes.
func (s *service) GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error {
	// if req.RaftAddr == "" || req.NodeID == "" {
	// 	return fmt.Errorf("not enough arguments: %+v", req)
	// }
	// return s.store.GlobalJoin(req.RaftAddr, req.NodeID)
	return nil
}

// GetUpdateNoti returns when the cmap is updated or timeout.
func (s *service) GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error {
	notiC := s.cmapAPI.GetUpdatedNoti(cmap.Version(req.Version))

	timeout := time.After(10 * time.Minute)
	for {
		select {
		case <-notiC:
			return nil
		case <-timeout:
			return errors.New("timeout, try again")
		}
	}
}
