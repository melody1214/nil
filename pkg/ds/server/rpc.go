package server

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/nilmux"
)

// Handler has exposed methods for rpc server.
type Handler struct {
	s *Server
}

func newNilRPCHandler(s *Server) (NilRPCHandler, error) {
	if s == nil {
		return nil, fmt.Errorf("nil server object")
	}

	return &Handler{s: s}, nil
}

// Hello is for testing rpc.
func (h *Handler) Hello(req *string, res *string) error {
	return h.s.handleHello(req, res)
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
	Hello(req *string, res *string) error
}

func (s *Server) handleHello(req *string, res *string) error {
	fmt.Println("Hello world")
	return nil
}
