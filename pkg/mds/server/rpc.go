package server

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
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

// NilRPCHandler is the interface of mds rpc commands.
type NilRPCHandler interface {
	// Join joins the mds node into the cluster.
	Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error
}

// Join joins the mds node into the cluster.
func (s *Server) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return s.store.Join(req.NodeID, req.RaftAddr)
}
