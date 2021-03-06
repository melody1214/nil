package cluster

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/ds/domain/model/device"
	"github.com/chanyoung/nil/app/ds/domain/model/volume"
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
	cfg              *config.Ds
	cmapAPI          cmap.SlaveAPI
	deviceRepository device.Repository
	volumeRepository volume.Repository
}

// NewService returns a new instance of a cluster map manager.
func NewService(cfg *config.Ds, cmapAPI cmap.SlaveAPI,
	deviceRepository device.Repository, volumeRepository volume.Repository) Service {
	logger = mlog.GetPackageLogger("app/ds/usecase/clustermap")

	return &service{
		cfg:              cfg,
		cmapAPI:          cmapAPI,
		deviceRepository: deviceRepository,
		volumeRepository: volumeRepository,
	}
}

// AddVolume adds a new volume with the given device path.
func (s *service) AddVolume(req *nilrpc.DCLAddVolumeRequest, res *nilrpc.DCLAddVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.AddVolume")

	// Get the current node ID.
	id, err := s.cmapAPI.ID()
	if err != nil {
		ctxLogger.Errorf("try to add volume %s, but the cmap is not initialized yet", req.DevicePath)
		return errors.Wrap(err, "can't add volume before cmap is initialized")
	}

	// Create a new device object.
	dev := device.New(device.Name(req.DevicePath))

	// Attach the device to the store.
	err = s.deviceRepository.Create(dev)
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	c := s.cmapAPI.SearchCall()
	node, err := c.Node().ID(id).Do()
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	// Update the volumes information to cmap.
	vols := s.volumeRepository.FindAll()
	node.Size = 0
	for _, v := range vols {
		node.Size = node.Size + (v.Size() / 1024 / 1024)
	}

	mds, err := c.Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(5*time.Second))
	if err != nil {
		ctxLogger.Error(err)
		return err
	}
	defer conn.Close()

	updateReq := &nilrpc.MMEUpdateNodeRequest{
		Node: node,
	}
	updateRes := &nilrpc.MMEUpdateNodeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsMembershipUpdateNode.String(), updateReq, updateRes); err != nil {
		ctxLogger.Error(err)
		return err
	}
	ctxLogger.Infof("add volume %s succeeded", req.DevicePath)

	return nil
}

// Service is the interface that provides rpc handlers.
type Service interface {
	AddVolume(req *nilrpc.DCLAddVolumeRequest, res *nilrpc.DCLAddVolumeResponse) error
	// RecoveryChunk(req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error
}
