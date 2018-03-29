package rpchandling

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// Join joins the mds node into the cluster.
func (h *Handler) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return h.store.Join(req.NodeID, req.RaftAddr)
}
