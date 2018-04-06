package admin

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/ds/delivery"
	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg   *config.Ds
	store Repository
	cMap  *cmap.CMap
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Ds, s Repository) delivery.AdminHandlers {
	logger = mlog.GetPackageLogger("app/ds/usecase/admin")

	return &handlers{
		cfg:   cfg,
		store: s,
		cMap:  cmap.New(),
	}
}

// AddVolume adds a new volume with the given device path.
func (h *handlers) AddVolume(req *nilrpc.AddVolumeRequest, res *nilrpc.AddVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.AddVolume")

	lv, err := repository.NewVol(req.DevicePath)
	if err != nil {
		return err
	}

	mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			ctxLogger.Error(err)
			return errors.Wrap(err, "failed to register volume")
		}
	}

	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	registerReq := &nilrpc.RegisterVolumeRequest{
		Ds:     h.cfg.ID,
		Size:   lv.Size,
		Free:   lv.Free,
		Used:   lv.Used,
		Speed:  lv.Speed.String(),
		Status: lv.Status.String(),
	}
	registerRes := &nilrpc.RegisterVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsAdminRegisterVolume.String(), registerReq, registerRes); err != nil {
		return err
	}

	lv.Name = registerRes.ID
	lv.MntPoint = "vol-" + lv.Name

	// 2) Add lv to the store service.
	if err := h.store.AddVolume(lv); err != nil {
		// TODO: remove added volume in the mds.
		return err
	}

	registerReq.ID = lv.Name
	registerReq.Size = lv.Size
	registerReq.Free = lv.Free
	registerReq.Used = lv.Used
	registerReq.Speed = lv.Speed.String()
	registerReq.Status = lv.Status.String()
	if err := cli.Call(nilrpc.MdsAdminRegisterVolume.String(), registerReq, registerRes); err != nil {
		// TODO: remove added volume in the mds and ds.
		return err
	}

	ctxLogger.Infof("add volume %s succeeded", lv.Name)
	return nil
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	h.cMap = m
}
