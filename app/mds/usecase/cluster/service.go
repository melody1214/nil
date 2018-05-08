package cluster

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg     *config.Mds
	store   Repository
	cmapAPI cmap.MasterAPI
}

// NewService creates a client service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.MasterAPI, s Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/cmapmap")

	return &service{
		cfg:     cfg,
		store:   s,
		cmapAPI: cmapAPI,
	}
}

// GetClusterMap returns a current local cmap.
func (s *service) GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error {
	res.ClusterMap = s.cmapAPI.GetLatestCMap()
	return nil
}

// GetUpdateNoti returns when the cmap is updated or timeout.
func (s *service) GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error {
	notiC := s.cmapAPI.GetUpdatedNoti(cmap.Version(req.Version))

	timeout := time.After(10 * time.Minute)
	for {
		select {
		case <-notiC:
			return nil
		case <-timeout:
			return errors.New("timeout, try again")
		}
	}
}

func (s *service) UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error {
	txid, err := s.store.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}

	if err = s.updateClusterMap(txid); err != nil {
		s.store.Rollback(txid)
		return err
	}
	if err = s.store.Commit(txid); err != nil {
		s.store.Rollback(txid)
		return err
	}

	s.rebalance()
	return nil
}

// LocalJoin handles the join request from the same local cluster nodes.
func (s *service) LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error {
	if s.canJoin(req.Node) == false {
		return fmt.Errorf("can't join into the cmap")
	}

	if err := s.store.LocalJoin(req.Node); err != nil {
		return errors.Wrap(err, "failed to add new node into the database")
	}

	return s.UpdateClusterMap(nil, nil)
}

// GlobalJoin handles the join request from the other raft nodes.
func (s *service) GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}
	return s.store.GlobalJoin(req.RaftAddr, req.NodeID)
}

func (s *service) canJoin(node cmap.Node) bool {
	// TODO: fill the checking rule.
	return true
}

func (s *service) rebalance() error {
	conn, err := nilrpc.Dial(s.cfg.ServerAddr+":"+s.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.MRERecoveryRequest{Type: nilrpc.Rebalance}
	res := &nilrpc.MRERecoveryResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.MdsRecoveryRecovery.String(), req, res)
}

// RegisterVolume receives a new volume information from ds and register it to the database.
func (s *service) RegisterVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	// If the id field of request is empty, then the ds
	// tries to get an id of volume.
	if req.ID == "" {
		return s.insertNewVolume(req, res)
	}
	return s.updateVolume(req, res)
}

func (s *service) updateVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateVolume")

	q := fmt.Sprintf(
		`
		UPDATE volume
		SET vl_status='%s', vl_size='%d', vl_free='%d', vl_used='%d', vl_max_encoding_group='%d', vl_speed='%s' 
		WHERE vl_id in ('%s')
		`, req.Status, req.Size, req.Free, req.Used, calcMaxChain(req.Size), req.Speed, req.ID,
	)

	_, err := s.store.Execute(repository.NotTx, q)
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	return s.UpdateClusterMap(nil, nil)
}

func (s *service) insertNewVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.insertNewVolume")

	q := fmt.Sprintf(
		`
		INSERT INTO volume (vl_node, vl_status, vl_size, vl_free, vl_used, vl_encoding_group, vl_max_encoding_group, vl_speed)
		SELECT node_id, '%s', '%d', '%d', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
		`, req.Status, req.Size, req.Free, req.Used, 0, calcMaxChain(req.Size), req.Speed, req.Ds,
	)

	r, err := s.store.Execute(repository.NotTx, q)
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		ctxLogger.Error(err)
		return err
	}
	res.ID = strconv.FormatInt(id, 10)

	return s.UpdateClusterMap(nil, nil)
}

func calcMaxChain(volumeSize uint64) int {
	if volumeSize <= 0 {
		return 0
	}

	// Test, chain per 10MB,
	return int(volumeSize / 10)
}

func (s *service) Stop() error {
	return s.store.Close()
}

// Service is the interface that provides clustermap domain's rpc handlers.
type Service interface {
	GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error
	GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error
	UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error
	RegisterVolume(req *nilrpc.MCLRegisterVolumeRequest, res *nilrpc.MCLRegisterVolumeResponse) error
	LocalJoin(req *nilrpc.MCLLocalJoinRequest, res *nilrpc.MCLLocalJoinResponse) error
	GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error
	JoinToGlobal(raftL *nilmux.Layer) error
	Stop() error
}
