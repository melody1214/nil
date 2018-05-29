package nilmux

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// NilMux is a default mux for nil communicatoins.
// Listen for tls tcp connection and handle it.
type NilMux struct {
	addr    string
	ln      net.Listener
	layers  []*Layer
	secuCfg *config.Security
}

// NewNilMux creates a NilMux object.
func NewNilMux(addr string, secuCfg *config.Security) *NilMux {
	logger = mlog.GetPackageLogger("pkg/nilmux")

	return &NilMux{
		addr:    addr,
		layers:  make([]*Layer, 0),
		secuCfg: secuCfg,
	}
}

// Address returns the listening address.
func (m *NilMux) Address() net.Addr {
	return m.ln.Addr()
}

// RegisterLayer regiters a layer to the NilMux.
func (m *NilMux) RegisterLayer(l *Layer) {
	m.layers = append(m.layers, l)
}

// Close closes the listener.
func (m *NilMux) Close() error {
	// Close real net.Listener first.
	// This will not accept more connections.
	if err := m.ln.Close(); err != nil {
		return err
	}

	// Close all registered layers.
	for _, l := range m.layers {
		if err := l.Close(); err != nil {
			return err
		}
	}

	return nil
}

// ListenAndServeTLS open a tls socket and route all incoming tcp connections.
func (m *NilMux) ListenAndServeTLS() error {
	ln, err := net.Listen("tcp", m.addr)
	if err != nil {
		return errors.Wrap(err, "NilMux ListenAndServeTLS failed")
	}

	// Make ssl cert.
	cert, err := tls.LoadX509KeyPair(
		m.secuCfg.CertsDir+"/"+m.secuCfg.ServerCrt,
		m.secuCfg.CertsDir+"/"+m.secuCfg.ServerKey,
	)
	if err != nil {
		return errors.Wrap(err, "NilMux ListenAndServeTLS failed")
	}

	// Load tls configuration and add certificate.
	tlsConfig := security.DefaultTLSConfig()
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, tlsConfig)
	go m.serve(tlsListener)
	return nil
}

func (m *NilMux) serve(ln net.Listener) error {
	if m.ln != nil {
		m.ln.Close()
	}
	m.ln = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go m.handleConn(conn)
	}
}

func (m *NilMux) handleConn(conn net.Conn) {
	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil {
		if err != io.EOF {
			logger.Errorf("failed to read the first byte: %v", err)
		}
		return
	}

	for _, l := range m.layers {
		if l.match(buf[0]) {
			l.handleConn(conn, buf[0])
			return
		}
	}

	// No matching layers.
	logger.Errorf("nilmux: no matching layers %+v\n", buf[0])
	conn.Close()
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted connections.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
