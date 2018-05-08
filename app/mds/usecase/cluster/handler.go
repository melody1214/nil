package cluster

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
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

// Join handles the join request from the other nodes.
func (s *service) Join(req *nilrpc.MCLJoinRequest, res *nilrpc.MCLJoinResponse) error {
	if s.canJoin(req.Node) == false {
		return fmt.Errorf("can't join into the cmap")
	}

	if err := s.store.JoinNewNode(req.Node); err != nil {
		return errors.Wrap(err, "failed to add new node into the database")
	}

	return s.UpdateClusterMap(nil, nil)
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

// Service is the interface that provides clustermap domain's rpc handlers.
type Service interface {
	GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error
	GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error
	UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error
	Join(req *nilrpc.MCLJoinRequest, res *nilrpc.MCLJoinResponse) error
}
