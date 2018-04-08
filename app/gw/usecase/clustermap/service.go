package clustermap

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Service manages the cluster map.
type Service struct {
	cMap *cmap.Controller
}

// NewService returns a new instance of a cluster map manager.
func NewService(cMap *cmap.Controller) *Service {
	logger = mlog.GetPackageLogger("app/gw/usecase/clustermap")

	return &Service{
		cMap: cMap,
	}
}

// Run starts to update cluster map periodically.
func (s *Service) Run() {
	go periodicUpdater(s.cMap)
	go realtimeUpdater(s.cMap)
}

func periodicUpdater(c *cmap.Controller) {
	ctxLogger := mlog.GetFunctionLogger(logger, "periodicUpdater")

	// Make ticker for routinely rebalancing.
	updateNoti := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-updateNoti.C:
			if err := c.Update(); err != nil {
				ctxLogger.Error(err)
			}
		}
	}
}

func realtimeUpdater(c *cmap.Controller) {
	ctxLogger := mlog.GetFunctionLogger(logger, "realtimeUpdater")

	for {
		mds, err := c.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to find alive mds"))
			time.Sleep(10 * time.Second)
			continue
		}

		if isUpdated(mds.Addr, c.LatestVersion()) {
			err = c.Update()
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to update cluster map"))
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func isUpdated(mds string, ver cmap.Version) bool {
	ctxLogger := mlog.GetFunctionLogger(logger, "isUpdated")

	conn, err := nilrpc.Dial(mds, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to dial to mds"))
		return false
	}
	defer conn.Close()

	req := &nilrpc.ClusterMapIsUpdatedRequest{Version: ver.Int64()}
	res := &nilrpc.ClusterMapIsUpdatedResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClustermapIsUpdated.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to talk with mds"))
		return false
	}
	defer cli.Close()

	return true
}
