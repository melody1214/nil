package cluster

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/mds/domain/model/region"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

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

// Join joins the node to the global cluster.
func (s *service) Join(raftL *nilmux.Layer) error {
	if err := s.rs.Open(raftL); err != nil {
		return err
	}

	// Need to join into the existing raft cluster.
	if s.cfg.Raft.LocalClusterAddr != s.cfg.Raft.GlobalClusterAddr {
		return raftJoin(
			s.cfg.Raft.GlobalClusterAddr,
			s.cfg.Raft.LocalClusterAddr,
			s.cfg.Raft.LocalClusterRegion,
		)
	}

	// I'm the first node of this cluster, no need to join.
	// Add my region into the region table.
	return s.rr.Create(&region.Region{
		Name:     region.Name(s.cfg.Raft.LocalClusterRegion),
		EndPoint: region.EndPoint(s.cfg.Raft.LocalClusterAddr),
	})
}

// Leave leaves the node from the global cluster.
func (s *service) Leave() error {
	return s.rs.Close()
}

// GlobalJoin handles the join request from the other raft nodes.
func (s *service) GlobalJoin(req *nilrpc.MCLGlobalJoinRequest, res *nilrpc.MCLGlobalJoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	if err := s.rs.Join(req.RaftAddr, req.NodeID); err != nil {
		return err
	}

	return s.rr.Create(&region.Region{
		Name:     region.Name(req.NodeID),
		EndPoint: region.EndPoint(req.RaftAddr),
	})
}
