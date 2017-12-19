package server

import (
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// MetaTransportLayer is a http.Transport layer which implements
// custom dial function for write rpcMeta header.
type MetaTransportLayer struct {
	addr    net.Addr
	connCh  chan net.Conn
	closed  uint32
	closeCh chan struct{}

	http.Transport
}

func newMetaTransportLayer(addr net.Addr) *MetaTransportLayer {
	return &MetaTransportLayer{
		addr:    addr,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),

		Transport: http.Transport{
			Dial: func(network, address string) (net.Conn, error) {
				defaultDialer := &net.Dialer{
					Timeout:   2 * time.Second,
					KeepAlive: 2 * time.Second,
				}

				conn, err := defaultDialer.Dial(network, address)
				if err != nil {
					return nil, err
				}

				_, err = conn.Write([]byte{byte(rpcMeta)})
				if err != nil {
					conn.Close()
					return nil, err
				}
				return conn, err
			},
			TLSHandshakeTimeout: 2 * time.Second,
		},
	}
}

// Addr returns the address of the meta transport layer.
func (mtl *MetaTransportLayer) Addr() net.Addr {
	return mtl.addr
}

// Accept waits and accepts the connection.
func (mtl *MetaTransportLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-mtl.connCh:
		return conn, nil
	case <-mtl.closeCh:
		return nil, errors.New("meta transport layer closed")
	}
}

// Close closes the meta transport layer.
func (mtl *MetaTransportLayer) Close() error {
	old := atomic.SwapUint32(&mtl.closed, closed)
	if old != closed {
		close(mtl.closeCh)
	}

	return nil
}

// GetHTTPClient returns http client which is implemented with custom dialer.
func (mtl *MetaTransportLayer) GetHTTPClient() http.Client {
	return http.Client{
		Transport: mtl,
	}
}

func (s *Server) handleMetaConn(conn net.Conn) {
	s.mtl.connCh <- conn
}
