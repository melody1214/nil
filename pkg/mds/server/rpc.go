package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/security"
)

func (s *Server) serveNilRPC(l *nilmux.Layer) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go s.nilRPCSrv.ServeConn(conn)
	}
}

func dialNilRPC(addr string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}

	config := security.DefaultTLSConfig()

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, config)
	if err != nil {
		return nil, err
	}

	// Write RPC header.
	_, err = conn.Write([]byte{
		0x02, // rpcNil
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, err
}

// JoinRequest includes an information for joining a new node into the raft clsuter.
// RaftAddr: address of the requested node.
// NodeID: ID of the requested node.
type JoinRequest struct {
	RaftAddr string
	NodeID   string
}

// JoinResponse is a NilRPC response message to join an existing cluster.
type JoinResponse struct{}

// NilRPCHandler implements the handlers for NilRPC requests.
type NilRPCHandler struct {
	// *rpc.Server
	s *Server
}

func newNilRPCHandler(s *Server) *NilRPCHandler {
	return &NilRPCHandler{
		s: s,
	}
}

// HandleJoin handles a raft cluster join request from remote.
func (n *NilRPCHandler) HandleJoin(req *JoinRequest, res *JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return n.s.store.Join(req.NodeID, req.RaftAddr)
}
