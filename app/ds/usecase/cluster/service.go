package cluster

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

// service manages the cluster map.
type service struct {
	cfg     *config.Ds
	cmapAPI cmap.SlaveAPI
	store   Repository
}

// NewService returns a new instance of a cluster map manager.
func NewService(cfg *config.Ds, cmapAPI cmap.SlaveAPI, s Repository) Service {
	logger = mlog.GetPackageLogger("app/ds/usecase/clustermap")

	return &service{
		cfg:     cfg,
		cmapAPI: cmapAPI,
		store:   s,
	}
}

// AddVolume adds a new volume with the given device path.
func (s *service) AddVolume(req *nilrpc.DCLAddVolumeRequest, res *nilrpc.DCLAddVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.AddVolume")

	lv, err := repository.NewVol(req.DevicePath)
	if err != nil {
		return err
	}

	mds, err := s.cmapAPI.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		ctxLogger.Error(err)
		return errors.Wrap(err, "failed to register volume")
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := s.cmapAPI.SearchCallNode().Name(cmap.NodeName(s.cfg.ID)).Do()
	if err != nil {
		return err
	}
	registerReq := &nilrpc.MCLRegisterVolumeRequest{
		Volume: cmap.Volume{
			Node:  n.ID,
			Stat:  cmap.VolumeStatus(lv.Status),
			Speed: cmap.VolumeSpeed(lv.Speed),
		},
	}
	registerRes := &nilrpc.MCLRegisterVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClusterRegisterVolume.String(), registerReq, registerRes); err != nil {
		return err
	}

	lv.Name = registerRes.ID
	lv.MntPoint = "vol-" + lv.Name
	if chunkSize, err := strconv.ParseInt(s.cfg.ChunkSize, 10, 64); err != nil {
		// Default 10MB.
		// TODO: make default config of volume.
		lv.ChunkSize = 10000000
	} else {
		lv.ChunkSize = chunkSize
	}

	// // 2) Add lv to the store service.
	// if err := s.store.AddVolume(lv); err != nil {
	// 	// TODO: remove added volume in the mds.
	// 	return err
	// }

	// registerReq.ID = lv.Name
	// registerReq.Size = lv.Size
	// registerReq.Free = lv.Free
	// registerReq.Used = lv.Used
	// registerReq.Speed = lv.Speed.String()
	// registerReq.Status = lv.Status.String()
	// if err := cli.Call(nilrpc.MdsClusterRegisterVolume.String(), registerReq, registerRes); err != nil {
	// 	// TODO: remove added volume in the mds and ds.
	// 	return err
	// }

	// go func() {
	// 	for {
	// 		ver := s.cmapAPI.GetLatestCMapVersion()
	// 		id, _ := strconv.ParseInt(lv.Name, 10, 64)
	// 		v, err := s.cmapAPI.SearchCallVolume().ID(cmap.ID(id)).Do()
	// 		if err != nil {
	// 			ctxLogger.Error(errors.Wrap(err, "failed to search volume, wait cmap to be updated"))
	// 			notiC := s.cmapAPI.GetUpdatedNoti(ver)
	// 			<-notiC
	// 			continue
	// 		}
	// 		v.Size = lv.Size
	// 		v.Stat = cmap.Active
	// 		if err = s.cmapAPI.UpdateVolume(v); err != nil {
	// 			ctxLogger.Error(errors.Wrap(err, "failed to update volume"))
	// 		}
	// 		break
	// 	}
	// }()

	ctxLogger.Infof("add volume %s succeeded", lv.Name)
	return nil
}

// Service is the interface that provides rpc handlers.
type Service interface {
	AddVolume(req *nilrpc.DCLAddVolumeRequest, res *nilrpc.DCLAddVolumeResponse) error
}
