package admin

import (
	"net/rpc"
	"strconv"
	"time"

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
	cfg     *config.Ds
	store   Repository
	cmapAPI cmap.SlaveAPI
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Ds, cmapAPI cmap.SlaveAPI, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/ds/usecase/admin")

	return &handlers{
		cfg:     cfg,
		store:   s,
		cmapAPI: cmapAPI,
	}
}

// AddVolume adds a new volume with the given device path.
func (h *handlers) AddVolume(req *nilrpc.DADAddVolumeRequest, res *nilrpc.DADAddVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.AddVolume")

	lv, err := repository.NewVol(req.DevicePath)
	if err != nil {
		return err
	}

	mds, err := h.cmapAPI.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		ctxLogger.Error(err)
		return errors.Wrap(err, "failed to register volume")
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	registerReq := &nilrpc.MADRegisterVolumeRequest{
		Ds:     h.cfg.ID,
		Size:   lv.Size,
		Free:   lv.Free,
		Used:   lv.Used,
		Speed:  lv.Speed.String(),
		Status: lv.Status.String(),
	}
	registerRes := &nilrpc.MADRegisterVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsAdminRegisterVolume.String(), registerReq, registerRes); err != nil {
		return err
	}

	lv.Name = registerRes.ID
	lv.MntPoint = "vol-" + lv.Name
	if chunkSize, err := strconv.ParseInt(h.cfg.ChunkSize, 10, 64); err != nil {
		// Default 10MB.
		// TODO: make default config of volume.
		lv.ChunkSize = 10000000
	} else {
		lv.ChunkSize = chunkSize
	}

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

	go func() {
		for {
			ver := h.cmapAPI.GetLatestCMapVersion()
			id, _ := strconv.ParseInt(lv.Name, 10, 64)
			v, err := h.cmapAPI.SearchCallVolume().ID(cmap.ID(id)).Do()
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to search volume, wait cmap to be updated"))
				notiC := h.cmapAPI.GetUpdatedNoti(ver)
				<-notiC
				continue
			}
			v.Size = lv.Size
			v.Stat = cmap.Active
			if err = h.cmapAPI.UpdateVolume(v); err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to update volume"))
			}
			break
		}
	}()

	ctxLogger.Infof("add volume %s succeeded", lv.Name)
	return nil
}

// Handlers is the interface that provides client http handlers.
type Handlers interface {
	AddVolume(req *nilrpc.DADAddVolumeRequest, res *nilrpc.DADAddVolumeResponse) error
}
