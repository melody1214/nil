package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/security"
)

func (s *Server) newNilRPCHandler() {
	s.NilRPCHandler = s
}

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

// NilRPCHandler is the interface of mds rpc commands.
type NilRPCHandler interface {
	// Join joins the mds node into the cluster.
	Join(req *JoinRequest, res *JoinResponse) error
}

// Join joins the mds node into the cluster.
func (s *Server) Join(req *JoinRequest, res *JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return s.store.Join(req.NodeID, req.RaftAddr)
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
