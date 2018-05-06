package clustermap

import (
	"time"

	"github.com/chanyoung/nil/pkg/cluster"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store      Repository
	clusterAPI cluster.MasterAPI
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(clusterAPI cluster.MasterAPI, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/clustermap")

	return &handlers{
		store:      s,
		clusterAPI: clusterAPI,
	}
}

// GetClusterMap returns a current local cluster map.
func (h *handlers) GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error {
	res.ClusterMap = h.clusterAPI.GetLatestCMap()
	return nil
}

// GetUpdateNoti returns when the cluster map is updated or timeout.
func (h *handlers) GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error {
	notiC := h.clusterAPI.GetUpdatedNoti(cluster.CMapVersion(req.Version))

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

func (h *handlers) UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error {
	txid, err := h.store.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}

	if err = h.updateClusterMap(txid); err != nil {
		h.store.Rollback(txid)
		return err
	}
	if err = h.store.Commit(txid); err != nil {
		h.store.Rollback(txid)
		return err
	}
	return nil
}

// Handlers is the interface that provides clustermap domain's rpc handlers.
type Handlers interface {
	GetClusterMap(req *nilrpc.MCLGetClusterMapRequest, res *nilrpc.MCLGetClusterMapResponse) error
	GetUpdateNoti(req *nilrpc.MCLGetUpdateNotiRequest, res *nilrpc.MCLGetUpdateNotiResponse) error
	UpdateClusterMap(req *nilrpc.MCLUpdateClusterMapRequest, res *nilrpc.MCLUpdateClusterMapResponse) error
}
