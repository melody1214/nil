package server

import (
	"errors"
	"net"
	"sync/atomic"
)

// GrpcTransportLayer implements a net.Listener interface.
type GrpcTransportLayer struct {
	addr    net.Addr
	connCh  chan net.Conn
	closed  uint32
	closeCh chan struct{}
}

func newGRPCTransportLayer(addr net.Addr) *GrpcTransportLayer {
	return &GrpcTransportLayer{
		addr:    addr,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),
	}
}

// Addr returns the address of the Grpc transport layer.
func (l *GrpcTransportLayer) Addr() net.Addr {
	return l.addr
}

// Accept waits and accepts the connection.
func (l *GrpcTransportLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closeCh:
		return nil, errors.New("grpc transport layer closed")
	}
}

// Close closes the grpc transport layer.
func (l *GrpcTransportLayer) Close() error {
	old := atomic.SwapUint32(&l.closed, closed)
	if old != closed {
		close(l.closeCh)
	}

	return nil
}

func (s *Server) handleGrpcConn(conn net.Conn) {
	// s.grpcTr.connCh <- conn
}
