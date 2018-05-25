package cmapmap

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

// Service manages the cmap.
type Service struct {
	cmapService *cmap.Service
}

// NewService returns a new instance of a cmap map manager.
func NewService(cmapService *cmap.Service) *Service {
	logger = mlog.GetPackageLogger("app/gw/usecase/cmapmap")

	return &Service{
		cmapService: cmapService,
	}
}

// Run starts to update cmap map periodically.
func (s *Service) Run() {
	go periodicUpdater(s.cmapService)
	go realtimeUpdater(s.cmapService)
}

func periodicUpdater(s *cmap.Service) {
	ctxLogger := mlog.GetFunctionLogger(logger, "periodicUpdater")

	// Make ticker for routinely updating.
	updateNoti := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-updateNoti.C:
			if err := updateClusterMap(s); err != nil {
				ctxLogger.Error(err)
			}
		}
	}
}

func realtimeUpdater(s *cmap.Service) {
	ctxLogger := mlog.GetFunctionLogger(logger, "realtimeUpdater")

	for {
		mds, err := s.SearchCall().Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to find alive mds"))
			time.Sleep(10 * time.Second)
			continue
		}

		if isUpdated(mds.Addr, s.GetLatestCMapVersion()) {
			if err := updateClusterMap(s); err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to update cmap"))
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func isUpdated(mdsAddr cmap.NodeAddress, ver cmap.Version) bool {
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
	if err := cli.Call(nilrpc.MdsClusterGetUpdateNoti.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to talk with mds"))
		return false
	}
	defer cli.Close()

	return true
}

func updateClusterMap(s *cmap.Service) error {
	mds, err := s.SearchCall().Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
	if err != nil {
		return err
	}

	cm, err := getLatestMapFromMDS(mds.Addr.String())
	if err != nil {
		return err
	}

	return s.UpdateCMap(cm)
}

func getLatestMapFromMDS(mdsAddr string) (*cmap.CMap, error) {
	conn, err := nilrpc.Dial(mdsAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &nilrpc.MCLGetClusterMapRequest{}
	res := &nilrpc.MCLGetClusterMapResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClusterGetClusterMap.String(), req, res); err != nil {
		return nil, err
	}

	return &res.ClusterMap, nil
}
