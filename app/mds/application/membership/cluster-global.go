package membership

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/mds/domain/model/region"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

func raftJoin(joinAddr, raftAddr, nodeID string) error {
	conn, err := nilrpc.Dial(joinAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.MMEGlobalJoinRequest{
		RaftAddr: raftAddr,
		NodeID:   nodeID,
	}

	res := &nilrpc.MMEGlobalJoinResponse{}

	cli := rpc.NewClient(conn)
	return cli.Call(nilrpc.MdsMembershipGlobalJoin.String(), req, res)
}

// Join joins the node to the global cluster.
func (s *service) Join(raftL *nilmux.Layer) error {
	ctxLogger := mlog.GetMethodLogger(logger, "service.Join")

	if err := s.rs.Open(raftL); err != nil {
		return err
	}

	// Set the open flag true.
	// It represents the database is now available.
	opened = true

	var err error
	// Need to join into the existing raft cluster.
	if s.cfg.Raft.LocalClusterAddr != s.cfg.Raft.GlobalClusterAddr {
		err = raftJoin(
			s.cfg.Raft.GlobalClusterAddr,
			s.cfg.Raft.LocalClusterAddr,
			s.cfg.Raft.LocalClusterRegion,
		)
	} else {
		// I'm the first node of this cluster, no need to join.
		// Add my region into the region table.
		err = s.rr.Create(&region.Region{
			Name:     region.Name(s.cfg.Raft.LocalClusterRegion),
			EndPoint: region.EndPoint(s.cfg.Raft.LocalClusterAddr),
		})
	}
	if err != nil {
		return err
	}

	// Initialize encoding matrices.
	if err := s.cr.InitEncodingMatricesID(); err != nil {
		ctxLogger.Fatal(err)
	}

	return nil
}

// Leave leaves the node from the global cluster.
func (s *service) Leave() error {
	return s.rs.Close()
}

// GlobalJoin handles the join request from the other raft nodes.
func (s *service) GlobalJoin(req *nilrpc.MMEGlobalJoinRequest, res *nilrpc.MMEGlobalJoinResponse) error {
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
