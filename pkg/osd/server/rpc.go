package server

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/nilmux"
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
	Hello(req *string, res *string) error
}

// Hello is for testing rpc.
func (s *Server) Hello(req *string, res *string) error {
	fmt.Println("Hello world")
	return nil
}
