package membership

import (
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg   *config.Mds
	cMap  *cmap.Controller
	store Repository

	swimSrv    *swim.Server
	swimTransL *nilmux.SwimTransportLayer
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, cMap *cmap.Controller, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/membership")

	return &handlers{
		cfg:   cfg,
		cMap:  cMap,
		store: s,
	}
}

func (h *handlers) GetMembershipList(req *nilrpc.GetMembershipListRequest, res *nilrpc.GetMembershipListResponse) error {
	res.Nodes = h.swimSrv.GetMap()
	return nil
}

// Create makes swim server.
func (h *handlers) Create(swimL *nilmux.Layer) (err error) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Create")

	h.swimTransL = nilmux.NewSwimTransportLayer(swimL)

	// Setup configuration.
	swimConf := swim.DefaultConfig()
	swimConf.ID = swim.ServerID(h.cfg.ID)
	swimConf.Address = swim.ServerAddress(h.cfg.ServerAddr + ":" + h.cfg.ServerPort)
	swimConf.Coordinator = swim.ServerAddress(h.cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(h.cfg.Swim.Period); err != nil {
		ctxLogger.Error(err)
	} else {
		swimConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(h.cfg.Swim.Expire); err != nil {
		ctxLogger.Error(err)
	} else {
		swimConf.PingExpire = t
	}
	swimConf.Type = swim.MDS

	h.swimSrv, err = swim.NewServer(swimConf, h.swimTransL)
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	return nil
}

// Run starts swim service.
func (h *handlers) Run() {
	sc := make(chan swim.PingError, 1)
	go h.swimSrv.Serve(sc)

	cmapUpdatedNotiC := h.cMap.GetUpdatedNoti(cmap.Version(0))
	for {
		select {
		case err := <-sc:
			h.recover(err)
			h.rebalance()
		case <-cmapUpdatedNotiC:
			latest := h.cMap.LatestVersion()
			h.swimSrv.SetCustomHeader("cmap_ver", strconv.FormatInt(latest.Int64(), 10))
			cmapUpdatedNotiC = h.cMap.GetUpdatedNoti(latest)
		}
	}
}

func (h *handlers) recover(pe swim.PingError) error {
	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.RecoverRequest{Pe: pe}
	res := &nilrpc.RecoverResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.MdsRecoveryRecover.String(), req, res)
}

func (h *handlers) rebalance() error {
	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.RebalanceRequest{}
	res := &nilrpc.RebalanceResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.MdsRecoveryRebalance.String(), req, res)
}

// Handlers is the interface that provides membership domain's rpc handlers.
type Handlers interface {
	GetMembershipList(req *nilrpc.GetMembershipListRequest, res *nilrpc.GetMembershipListResponse) error
	Create(swimL *nilmux.Layer) error
	Run()
}
