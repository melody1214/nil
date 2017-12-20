package server

import (
	"crypto/tls"
	"io"
	"net"

	"github.com/chanyoung/nil/pkg/security"
)

// TrSign is the first byte of income data that implies the type of connection.
type TrSign byte

const (
	// TrRaft is the sign of raft rpc connection.
	// All core communication traffic from the hashicorp raft library
	// has the TrSign byte at the first byte of the data.
	TrRaft TrSign = 0x01

	// The rest of connection types are handled by http mux.
)

// listenAndServeTLS open a tls socket and route all incoming tcp connections.
func (s *Server) listenAndServeTLS() {
	// Make ssl cert.
	cert, err := tls.LoadX509KeyPair(
		s.cfg.Security.CertsDir+"/"+s.cfg.Security.ServerCrt,
		s.cfg.Security.CertsDir+"/"+s.cfg.Security.ServerKey,
	)
	if err != nil {
		log.Error(err)
		return
	}

	// Load tls configuration and add certificate.
	tlsConfig := security.DefaultTLSConfig()
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	// Start listen and serve.
	ln, err := tls.Listen("tcp", ":"+s.cfg.ServerPort, tlsConfig)
	if err != nil {
		log.Error(err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		go s.handleConn(conn)
	}
}

// handleConn do the primary classification by checking TrSign.
func (s *Server) handleConn(conn net.Conn) {
	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil {
		if err != io.EOF {
			log.Errorf("failed to read a rpc header byte: %v", err)
		}
		return
	}

	switch TrSign(buf[0]) {
	case TrRaft:
		s.handleRaftConn(conn)

	default:
		s.handleHTTPConn(conn)
	}
}
