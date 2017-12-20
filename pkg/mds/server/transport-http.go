package server

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
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

func (s *Server) handleHTTPConn(conn net.Conn) {
	s.httpTr.connCh <- conn
}
