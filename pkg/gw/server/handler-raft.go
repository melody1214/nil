package server

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/security"
)

// handleRaftConn simply resend the connection to the remote mds.
func (s *Server) handleRaftConn(conn net.Conn) {
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	tlsConfig := security.DefaultTLSConfig()

	// Temporarily hard fixed to first mds.
	remote, err := tls.DialWithDialer(dialer, "tcp", s.cfg.FirstMds, tlsConfig)
	if err != nil {
		log.Error(err)
		return
	}

	// Resign the TrRaft sign to remote mds.
	_, err = remote.Write([]byte{byte(TrRaft)})
	if err != nil {
		log.Error(err)
		return
	}
	go io.Copy(conn, remote)
	go io.Copy(remote, conn)
}
