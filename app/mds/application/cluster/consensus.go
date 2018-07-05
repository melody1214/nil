package cluster

import (
	"net/rpc"
	"time"

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
