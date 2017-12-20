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
	l := &RaftTransportLayer{
		addr:    advertise,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),
	}
	return l
}

// Addr returns the address of the raft transport layer.
func (l *RaftTransportLayer) Addr() net.Addr {
	return l.addr
}

// Accept waits and accepts the connection.
func (l *RaftTransportLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closeCh:
		return nil, errors.New("raft transport layer closed")
	}
}

// Close closes the raft transport layer.
func (l *RaftTransportLayer) Close() error {
	old := atomic.SwapUint32(&l.closed, closed)
	if old != closed {
		close(l.closeCh)
	}

	return nil
}

// Dial dials to the given address and returns a new network connection.
func (l *RaftTransportLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}

	config := security.DefaultTLSConfig()

	conn, err := tls.DialWithDialer(dialer, "tcp", string(addr), config)
	if err != nil {
		return nil, err
	}

	// Write RPC header.
	_, err = conn.Write([]byte{byte(TrRaft)})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, err
}

func (s *Server) handleRaftConn(conn net.Conn) {
	s.raftTr.connCh <- conn
}
