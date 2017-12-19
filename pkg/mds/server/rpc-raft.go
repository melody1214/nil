package server

import (
	"crypto/tls"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/chanyoung/nil/pkg/security"
	"github.com/hashicorp/raft"
)

const closed uint32 = 1

// RaftTransportLayer is the implementation of the raft.StreamLayer.
type RaftTransportLayer struct {
	addr    net.Addr
	connCh  chan net.Conn
	closed  uint32
	closeCh chan struct{}
}

func newRaftTransportLayer(advertise net.Addr) *RaftTransportLayer {
	rtl := &RaftTransportLayer{
		addr:    advertise,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),
	}
	return rtl
}

// Addr returns the address of the raft transport layer.
func (rtl *RaftTransportLayer) Addr() net.Addr {
	return rtl.addr
}

// Accept waits and accepts the connection.
func (rtl *RaftTransportLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-rtl.connCh:
		return conn, nil
	case <-rtl.closeCh:
		return nil, errors.New("raft transport layer closed")
	}
}

// Close closes the raft transport layer.
func (rtl *RaftTransportLayer) Close() error {
	old := atomic.SwapUint32(&rtl.closed, closed)
	if old != closed {
		close(rtl.closeCh)
	}

	return nil
}

// Dial dials to the given address and returns a new network connection.
func (rtl *RaftTransportLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}

	config := security.DefaultTLSConfig()

	conn, err := tls.DialWithDialer(dialer, "tcp", string(addr), config)
	if err != nil {
		return nil, err
	}

	// Write RPC header.
	_, err = conn.Write([]byte{byte(rpcRaft)})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, err
}

func (s *Server) handleRaftConn(conn net.Conn) {
	s.rtl.connCh <- conn
}
