package server

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const closed uint32 = 1

type httpTransportLayer struct {
	addr    net.Addr
	connCh  chan net.Conn
	closed  uint32
	closeCh chan struct{}
}

func newHTTPTransportLayer(addr net.Addr) *httpTransportLayer {
	return &httpTransportLayer{
		addr:    addr,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),
	}
}

// Addr returns the address of the http transport layer.
func (l *httpTransportLayer) Addr() net.Addr {
	return l.addr
}

// Accept waits and accepts the connection.
func (l *httpTransportLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closeCh:
		return nil, errors.New("http transport layer closed")
	}
}

// Close closes the http transport layer.
func (l *httpTransportLayer) Close() error {
	old := atomic.SwapUint32(&l.closed, closed)
	if old != closed {
		close(l.closeCh)
	}

	return nil
}

func (s *Server) handleHTTPConn(conn net.Conn, signByte byte) {
	s.httpTr.connCh <- newNilConn(conn, signByte)
}

type nilConn struct {
	conn     net.Conn
	once     sync.Once
	signByte byte
}

func newNilConn(conn net.Conn, signByte byte) *nilConn {
	return &nilConn{
		conn:     conn,
		signByte: signByte,
	}
}

func (nc *nilConn) Read(b []byte) (n int, err error) {
	nc.once.Do(func() {
		if len(b) < 1 {
			return
		}

		b[0] = nc.signByte
		b = b[1:]
		n++
	})
	read, err := nc.conn.Read(b)
	return read + n, err
}

func (nc *nilConn) Write(b []byte) (n int, err error) {
	return nc.conn.Write(b)
}

func (nc *nilConn) Close() error {
	return nc.conn.Close()
}

func (nc *nilConn) LocalAddr() net.Addr {
	return nc.conn.LocalAddr()
}

func (nc *nilConn) RemoteAddr() net.Addr {
	return nc.conn.RemoteAddr()
}

func (nc *nilConn) SetDeadline(t time.Time) error {
	return nc.conn.SetDeadline(t)
}

func (nc *nilConn) SetReadDeadline(t time.Time) error {
	return nc.conn.SetReadDeadline(t)
}

func (nc *nilConn) SetWriteDeadline(t time.Time) error {
	return nc.conn.SetWriteDeadline(t)
}
