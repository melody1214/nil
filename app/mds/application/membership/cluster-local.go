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
	ctxLogger := mlog.GetMethodLogger(logger, "service.LocalJoin")

	if err := s.updateNode(&req.Node); err != nil {
		ctxLogger.Infof("node from %s failed to join into the local cluster by error %v", req.Node.Addr.String(), err)
		return err
	}

	ctxLogger.Infof("node from %s succeed to join into the local cluster", req.Node.Addr.String())
	return nil
}

// UpdateNode handles the update node request from the same local cluster nodes.
func (s *service) UpdateNode(req *nilrpc.MMEUpdateNodeRequest, res *nilrpc.MMEUpdateNodeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "service.UpdateNode")

	if err := s.updateNode(&req.Node); err != nil {
		ctxLogger.Infof("node from %s failed to updated by error %v", req.Node.Addr.String(), err)
		return err
	}

	ctxLogger.Infof("node from %s succeed to update itself", req.Node.Addr.String())
	return nil
}

func (s *service) updateNode(node *cmap.Node) error {
	if !opened {
		return fmt.Errorf("database is not opened yet")
	}

	updated, err := s.cr.UpdateNode(node)
	if err != nil {
		return err
	}

	s.cmapAPI.UpdateCMap(updated)
	go s.rebalance()

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
