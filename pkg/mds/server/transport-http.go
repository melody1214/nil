package server

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chanyoung/nil/pkg/security"
)

// HTTPTransportLayer is a http.Transport layer which implements
// custom dial function for write rpcMeta header.
type HTTPTransportLayer struct {
	addr    net.Addr
	connCh  chan net.Conn
	closed  uint32
	closeCh chan struct{}

	http.Transport
}

func newHTTPTransportLayer(addr net.Addr) *HTTPTransportLayer {
	return &HTTPTransportLayer{
		addr:    addr,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),

		Transport: http.Transport{
			Dial: func(network, address string) (net.Conn, error) {
				dialer := &net.Dialer{Timeout: 2 * time.Second}

				config := security.DefaultTLSConfig()

				conn, err := tls.DialWithDialer(dialer, "tcp", address, config)
				if err != nil {
					return nil, err
				}

				// _, err = conn.Write([]byte{byte(TrR)})
				// if err != nil {
				// 	conn.Close()
				// 	return nil, err
				// }
				return conn, err
			},
			TLSHandshakeTimeout: 2 * time.Second,
		},
	}
}

// Addr returns the address of the meta transport layer.
func (l *HTTPTransportLayer) Addr() net.Addr {
	return l.addr
}

// Accept waits and accepts the connection.
func (l *HTTPTransportLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closeCh:
		return nil, errors.New("meta transport layer closed")
	}
}

// Close closes the meta transport layer.
func (l *HTTPTransportLayer) Close() error {
	old := atomic.SwapUint32(&l.closed, closed)
	if old != closed {
		close(l.closeCh)
	}

	return nil
}

// GetHTTPClient returns http client which is implemented with custom dialer.
func (l *HTTPTransportLayer) GetHTTPClient() http.Client {
	return http.Client{
		Transport: l,
	}
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
