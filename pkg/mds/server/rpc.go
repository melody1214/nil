package server

import (
	"crypto/tls"
	"io"
	"net"

	"github.com/chanyoung/nil/pkg/security"
)

// RPCHeader is the first byte of tcp connection that implies the type of RPC call.
type RPCHeader byte

const (
	rpcRaft  RPCHeader = 0x01
	rpcSwim            = 0x02
	rpcMeta            = 0x03
	rpcAdmin           = 0x04
)

func (s *Server) listenAndServeTLS() {
	cert, err := tls.LoadX509KeyPair(
		s.cfg.Security.CertsDir+"/"+s.cfg.Security.ServerCrt,
		s.cfg.Security.CertsDir+"/"+s.cfg.Security.ServerKey,
	)
	if err != nil {
		log.Error(err)
		return
	}

	config := security.DefaultTLSConfig()
	config.Certificates = append(config.Certificates, cert)
	ln, err := tls.Listen("tcp", ":"+s.cfg.ServerPort, config)
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

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil {
		if err != io.EOF {
			log.Error("failed to read a rpc header byte: %v", err)
		}
		return
	}

	switch RPCHeader(buf[0]) {
	case rpcRaft:
		s.handleRaftConn(conn)

	case rpcSwim:

	case rpcMeta:

	case rpcAdmin:

	default:
		log.Error("unknown rpc header: %v", buf[0])
		return
	}
}
