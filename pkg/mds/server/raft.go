package server

import (
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/hashicorp/raft"
)

type raftTransportLayer struct {
	*nilmux.Layer
}

func newRaftTransportLayer(l *nilmux.Layer) *raftTransportLayer {
	return &raftTransportLayer{Layer: l}
}

func (l *raftTransportLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return nilrpc.Dial(string(addr), nilrpc.RPCRaft, timeout)
}
