package clustermap

import (
	"github.com/chanyoung/nil/pkg/cluster"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Service manages the cluster map.
type Service struct {
	clusterAPI cluster.SlaveAPI
}

// NewService returns a new instance of a cluster map manager.
func NewService(clusterAPI cluster.SlaveAPI) *Service {
	logger = mlog.GetPackageLogger("app/ds/usecase/clustermap")

	return &Service{
		clusterAPI: clusterAPI,
	}
}

// // Run starts to update cluster map periodically.
// func (s *Service) Run() {
// 	go updater(s.clusterAPI)
// }

// func updater(c cluster.SlaveAPI) {
// 	ctxLogger := mlog.GetFunctionLogger(logger, "Service.updater")

// 	// Make ticker for routinely rebalancing.
// 	updateNoti := time.NewTicker(30 * time.Second)

// 	// outdatedC := c.GetOutdatedNoti()
// 	for {
// 		select {
// 		case <-updateNoti.C:
// 			if err := updateClusterMap(c); err != nil {
// 				ctxLogger.Error(err)
// 			}
// 		case <-outdatedC:
// 			outdatedC = c.GetOutdatedNoti()
// 			if err := updateClusterMap(c); err != nil {
// 				ctxLogger.Error(err)
// 			}
// 		}
// 	}
// }

// func updateClusterMap(c *cmap.Controller) error {
// 	mds, err := c.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
// 	if err != nil {
// 		return err
// 	}

// 	cm, err := getLatestMapFromMDS(mds.Addr)
// 	if err != nil {
// 		return err
// 	}

// 	return c.Update(cm)
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
