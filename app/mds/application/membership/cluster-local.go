package membership

import (
	"errors"
	"fmt"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

// GetClusterMap returns a current local cmap.
func (s *service) GetClusterMap(req *nilrpc.MMEGetClusterMapRequest, res *nilrpc.MMEGetClusterMapResponse) error {
	res.ClusterMap = s.cmapAPI.GetLatestCMap()
	return nil
}

// LocalJoin handles the join request from the same local cluster nodes.
func (s *service) LocalJoin(req *nilrpc.MMELocalJoinRequest, res *nilrpc.MMELocalJoinResponse) error {
	if !opened {
		return fmt.Errorf("database is not opened yet")
	}

	ctxLogger := mlog.GetMethodLogger(logger, "service.LocalJoin")
	ctxLogger.Infof("node %v try to join into the local cluster", req.Node)

	updated, err := s.cr.UpdateNode(&req.Node)
	if err != nil {
		ctxLogger.Infof("node %v failed to join into the local cluster by error %v", req.Node, err)
		return err
	}

	ctxLogger.Infof("%+v", updated)

	s.cmapAPI.UpdateCMap(updated)
	ctxLogger.Infof("node %v succeed to join into the local cluster", req.Node)
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
