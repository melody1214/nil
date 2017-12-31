package server

import (
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

type swimTransportLayer struct {
	*nilmux.Layer
}

func newSwimTransportLayer(l *nilmux.Layer) *swimTransportLayer {
	return &swimTransportLayer{Layer: l}
}

func (l *swimTransportLayer) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	return nilrpc.Dial(string(addr), nilrpc.RPCSwim, timeout)
}
