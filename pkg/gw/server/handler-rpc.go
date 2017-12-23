package server

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/security"
)

// TODO: conn close?
func (s *Server) serveNil(l *nilmux.Layer) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		dialer := &net.Dialer{Timeout: 2 * time.Second}
		tlsConfig := security.DefaultTLSConfig()

		// Temporarily hard fixed to first mds.
		remote, err := tls.DialWithDialer(dialer, "tcp", s.cfg.FirstMds, tlsConfig)
		if err != nil {
			log.Error(err)
			return
		}

		go io.Copy(conn, remote)
		go io.Copy(remote, conn)
	}
}
