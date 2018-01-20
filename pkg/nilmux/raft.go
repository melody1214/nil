package nilmux

import (
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/hashicorp/raft"
)

// RaftTransportLayer provides dial function with RPCRaft byte.
type RaftTransportLayer struct {
	*Layer
}

// NewRaftTransportLayer returns RaftTransportLayer on the top of the given layer.
func NewRaftTransportLayer(l *Layer) *RaftTransportLayer {
	return &RaftTransportLayer{Layer: l}
}

// Dial dials rpc with RPCRaft byte.
func (l *RaftTransportLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return nilrpc.Dial(string(addr), nilrpc.RPCRaft, timeout)
}
