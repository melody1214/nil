package membership

import (
	"errors"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetClusterMap returns a current local cmap.
func (s *service) GetClusterMap(req *nilrpc.MMEGetClusterMapRequest, res *nilrpc.MMEGetClusterMapResponse) error {
	res.ClusterMap = s.cmapAPI.GetLatestCMap()
	return nil
}

// LocalJoin handles the join request from the same local cluster nodes.
func (s *service) LocalJoin(req *nilrpc.MMELocalJoinRequest, res *nilrpc.MMELocalJoinResponse) error {
	return nil
}

// GetUpdateNoti returns when the cmap is updated or timeout.
func (s *service) GetUpdateNoti(req *nilrpc.MMEGetUpdateNotiRequest, res *nilrpc.MMEGetUpdateNotiResponse) error {
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
