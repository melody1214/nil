package clustermap

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cluster"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Service manages the cluster map.
type Service struct {
	clusterAPI cluster.SlaveAPI
}

// NewService returns a new instance of a cluster map manager.
func NewService(clusterAPI cluster.SlaveAPI) *Service {
	logger = mlog.GetPackageLogger("app/gw/usecase/clustermap")

	return &Service{
		clusterAPI: clusterAPI,
	}
}

// Run starts to update cluster map periodically.
func (s *Service) Run() {
	go periodicUpdater(s.clusterAPI)
	go realtimeUpdater(s.clusterAPI)
}

func periodicUpdater(c cluster.SlaveAPI) {
	ctxLogger := mlog.GetFunctionLogger(logger, "periodicUpdater")

	// Make ticker for routinely rebalancing.
	updateNoti := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-updateNoti.C:
			if err := c.UpdateFromMDS(); err != nil {
				ctxLogger.Error(err)
			}
		}
	}
}

func realtimeUpdater(c cluster.SlaveAPI) {
	ctxLogger := mlog.GetFunctionLogger(logger, "realtimeUpdater")

	for {
		mds, err := c.SearchCallNode().Type(cluster.MDS).Status(cluster.Alive).Do()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to find alive mds"))
			time.Sleep(10 * time.Second)
			continue
		}

		if isUpdated(mds.Addr, c.GetLatestCMapVersion()) {
			err := c.UpdateFromMDS()
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to update cluster map"))
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func isUpdated(mdsAddr cluster.NodeAddress, ver cluster.CMapVersion) bool {
	ctxLogger := mlog.GetFunctionLogger(logger, "isUpdated")

	conn, err := nilrpc.Dial(mdsAddr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to dial to mds"))
		return false
	}
	defer conn.Close()

	req := &nilrpc.MCLGetUpdateNotiRequest{Version: ver.Int64()}
	res := &nilrpc.MCLGetUpdateNotiResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClustermapGetUpdateNoti.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to talk with mds"))
		return false
	}
	defer cli.Close()

	return true
}

// func updateClusterMap(c cluster.SlaveAPI) error {
// 	mds, err := c.SearchCallNode().Type(cluster.MDS).Status(cluster.Alive).Do()
// 	if err != nil {
// 		return err
// 	}

// 	cm, err := getLatestMapFromMDS(mds.Addr.String())
// 	if err != nil {
// 		return err
// 	}

// 	return c.(cm)
// }

// func getLatestMapFromMDS(mdsAddr string) (*cmap.CMap, error) {
// 	conn, err := nilrpc.Dial(mdsAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer conn.Close()

// 	req := &nilrpc.MCLGetClusterMapRequest{}
// 	res := &nilrpc.MCLGetClusterMapResponse{}

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.MdsClustermapGetClusterMap.String(), req, res); err != nil {
// 		return nil, err
// 	}

// 	return &res.ClusterMap, nil
// }
