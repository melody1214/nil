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

	const (
		hot  = uint(5)
		cold = uint(1)
	)

	vol, err := repository.NewVol(req.DevicePath, hot, cold)
	if err != nil {
		return err
	}

	c := s.cmapAPI.SearchCall()
	mds, err := c.Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
	if err != nil {
		ctxLogger.Error(err)
		return errors.Wrap(err, "failed to register volume")
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := c.Node().Name(cmap.NodeName(s.cfg.ID)).Do()
	if err != nil {
		return err
	}
	registerReq := &nilrpc.MCLRegisterVolumeRequest{
		Volume: cmap.Volume{
			Node:  n.ID,
			Stat:  cmap.VolumeStatus(vol.Status),
			Speed: cmap.VolumeSpeed(vol.Speed),
		},
	}
	registerRes := &nilrpc.MCLRegisterVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClusterRegisterVolume.String(), registerReq, registerRes); err != nil {
		return err
	}

	vol.Name = registerRes.ID.String()
	vol.MntPoint = "vol-" + vol.Name
	if chunkSize, err := strconv.ParseInt(s.cfg.ChunkSize, 10, 64); err != nil {
		// Default 10MB.
		// TODO: make default config of volume.
		vol.ChunkSize = 10000000
	} else {
		vol.ChunkSize = chunkSize
	}

	// 2) Add lv to the store service.
	if err := s.store.AddVolume(vol); err != nil {
		// TODO: remove added volume in the mds.
		return err
	}

	go func() {
		for {
			c := s.cmapAPI.SearchCall()
			v, err := c.Volume().ID(registerRes.ID).Do()
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to search volume, wait cmap to be updated"))
				notiC := s.cmapAPI.GetUpdatedNoti(c.Version())
				<-notiC
				continue
			}
			v.Size = vol.Size
			v.Speed = cmap.VolumeSpeed(vol.Speed.String())
			v.Stat = cmap.VolumeStatus(vol.Status.String())
			if err = s.cmapAPI.UpdateVolume(v); err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to update volume"))
			}
			break
		}
	}()

	ctxLogger.Infof("add volume %s succeeded", vol.Name)
	return nil
}

// Service is the interface that provides rpc handlers.
type Service interface {
	AddVolume(req *nilrpc.DCLAddVolumeRequest, res *nilrpc.DCLAddVolumeResponse) error
}
