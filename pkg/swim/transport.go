package swim

import (
	"net"
	"time"
)

// Transport is Swim network transport abstraction layer.
type Transport interface {
	net.Listener

	Dial(address string, timeout time.Duration) (net.Conn, error)
}
