package nilmux

import (
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// SwimTransportLayer provides dial function with RPCSwim byte.
type SwimTransportLayer struct {
	*Layer
}

// NewSwimTransportLayer returns SwimTransportLayer on the top of the given layer.
func NewSwimTransportLayer(l *Layer) *SwimTransportLayer {
	return &SwimTransportLayer{Layer: l}
}

// Dial dials rpc with RPCSwim byte.
func (l *SwimTransportLayer) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	return nilrpc.Dial(string(addr), nilrpc.RPCSwim, timeout)
}
