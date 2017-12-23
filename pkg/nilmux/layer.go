package nilmux

import (
	"net"
	"sync/atomic"

	"github.com/pkg/errors"
)

const closed uint32 = 1

// Layer is NilMux rpc layer.
type Layer struct {
	rpcTypes            []byte
	preserveRPCTypeByte bool

	addr    net.Addr
	connCh  chan net.Conn
	closed  uint32
	closeCh chan struct{}
}

// NewLayer makes a transport layer with the given rpcType byte.
func NewLayer(rpcTypes []byte, advertise net.Addr, preserveRPCTypeByte bool) *Layer {
	return &Layer{
		rpcTypes:            rpcTypes,
		preserveRPCTypeByte: preserveRPCTypeByte,
		addr:                advertise,
		connCh:              make(chan net.Conn),
		closeCh:             make(chan struct{}),
	}
}

func (l *Layer) match(b byte) bool {
	for _, rpcType := range l.rpcTypes {
		if rpcType == b {
			return true
		}
	}
	return false
}

// Addr returns the address of the transport layer.
func (l *Layer) Addr() net.Addr {
	return l.addr
}

// Accept waits and accepts the connection.
func (l *Layer) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closeCh:
		return nil, errors.New("transport layer closed")
	}
}

// Close closes the transport layer.
func (l *Layer) Close() error {
	old := atomic.SwapUint32(&l.closed, closed)
	if old != closed {
		close(l.closeCh)
	}

	return nil
}

func (l *Layer) handleConn(conn net.Conn, rpcType byte) {
	if l.preserveRPCTypeByte {
		l.connCh <- newNilConn(conn, rpcType)
	} else {
		l.connCh <- conn
	}
}
