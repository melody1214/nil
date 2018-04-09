package consensus

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg   *config.Mds
	store Repository
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/consensus")

	return &handlers{
		cfg:   cfg,
		store: s,
	}
}

func (h *handlers) Open(raftL *nilmux.Layer) error {
	return h.store.Open(raftL)
}

func (h *handlers) Stop() error {
	return h.store.Close()
}

// Join joins into the existing cluster, located at joinAddr.
// The joinAddr node must the leader state node.
func (h *handlers) Join() error {
	if h.cfg.Raft.LocalClusterAddr != h.cfg.Raft.GlobalClusterAddr {
		return join(h.cfg.Raft.GlobalClusterAddr, h.cfg.Raft.LocalClusterAddr, h.cfg.Raft.LocalClusterRegion)
	}

	return nil
}

func join(joinAddr, raftAddr, nodeID string) error {
	conn, err := nilrpc.Dial(joinAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.JoinRequest{
		RaftAddr: raftAddr,
		NodeID:   nodeID,
	}

	res := &nilrpc.JoinResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.MdsAdminJoin.String(), req, res)
}

// Handlers is the interface that provides consensus domain's rpc handlers.
type Handlers interface {
	Open(raftL *nilmux.Layer) error
	Stop() error
	Join() error
}
