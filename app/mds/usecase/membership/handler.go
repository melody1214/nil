package membership

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg         *config.Mds
	cmapService *cmap.Service
	store       Repository
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, cmapService *cmap.Service, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/membership")

	return &handlers{
		cfg:         cfg,
		cmapService: cmapService,
		store:       s,
	}
}

// func (h *handlers) GetMembershipList(req *nilrpc.GetMembershipListRequest, res *nilrpc.GetMembershipListResponse) error {
// 	res.Nodes = h.cmapService.GetMap()
// 	return nil
// }

// Setup setup the swim server and run.
func (h *handlers) Setup(swimL *nilmux.Layer) (err error) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Create")

	// Setup configuration.
	cmapConf := cmap.DefaultConfig()
	cmapConf.Name = cmap.NodeName(h.cfg.ID)
	cmapConf.Address = cmap.NodeAddress(h.cfg.ServerAddr + ":" + h.cfg.ServerPort)
	cmapConf.Coordinator = cmap.NodeAddress(h.cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(h.cfg.Swim.Period); err != nil {
		ctxLogger.Error(err)
	} else {
		cmapConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(h.cfg.Swim.Expire); err != nil {
		ctxLogger.Error(err)
	} else {
		cmapConf.PingExpire = t
	}
	cmapConf.Type = cmap.MDS

	h.cmapService.StartMembershipServer(*cmapConf, nilmux.NewSwimTransportLayer(swimL))
	return nil
}

// // Run starts swim service.
// func (h *handlers) Run() {
// 	sc := make(chan swim.PingError, 1)
// 	go h.swimSrv.Serve(sc)

// 	cmapUpdatedNotiC := h.cmapAPI.GetUpdatedNoti(cmap.Version(0))
// 	for {
// 		select {
// 		case err := <-sc:
// 			if err.Err == swim.ErrChanged.Error() {
// 				// If the error message said it is occured by new member join,
// 				// then do rebalance and finish.
// 				h.recover()
// 			}
// 			h.rebalance()
// 		case <-cmapUpdatedNotiC:
// 			// TODO: redundant mechanism with the above swim error channel?
// 			h.rebalance()
// 			latest := h.cmapAPI.LatestVersion()
// 			h.swimSrv.SetCustomHeader("cmap_ver", strconv.FormatInt(latest.Int64(), 10))
// 			cmapUpdatedNotiC = h.cMap.GetUpdatedNoti(latest)
// 		}
// 	}
// }

// func (h *handlers) recover() error {
// 	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	req := &nilrpc.MRERecoveryRequest{Type: nilrpc.Recover}
// 	res := &nilrpc.MRERecoveryResponse{}

// 	cli := rpc.NewClient(conn)
// 	return cli.Call(nilrpc.MdsRecoveryRecovery.String(), req, res)
// }

// func (h *handlers) rebalance() error {
// 	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	req := &nilrpc.MRERecoveryRequest{Type: nilrpc.Rebalance}
// 	res := &nilrpc.MRERecoveryResponse{}

// 	cli := rpc.NewClient(conn)
// 	return cli.Call(nilrpc.MdsRecoveryRecovery.String(), req, res)
// }

// Handlers is the interface that provides membership domain's rpc handlers.
type Handlers interface {
	// GetMembershipList(req *nilrpc.GetMembershipListRequest, res *nilrpc.GetMembershipListResponse) error
	// Create(swimL *nilmux.Layer) error
	Setup(swimL *nilmux.Layer) error
}
