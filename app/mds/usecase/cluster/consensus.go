package cluster

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

func (s *service) JoinToGlobal(raftL *nilmux.Layer) error {
	if err := s.store.Open(raftL); err != nil {
		return err
	}

	// I'm the first node of this cluster, no need to join.
	if s.cfg.Raft.LocalClusterAddr == s.cfg.Raft.GlobalClusterAddr {
		return nil
	}

	return raftJoin(s.cfg.Raft.GlobalClusterAddr, s.cfg.Raft.LocalClusterAddr, s.cfg.Raft.LocalClusterRegion)
}

func raftJoin(joinAddr, raftAddr, nodeID string) error {
	conn, err := nilrpc.Dial(joinAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.MCLGlobalJoinRequest{
		RaftAddr: raftAddr,
		NodeID:   nodeID,
	}

	res := &nilrpc.MCLGlobalJoinResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.MdsClusterGlobalJoin.String(), req, res)
}
