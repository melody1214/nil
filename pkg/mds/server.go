package mds

import (
	"net"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// server serve RPCs.
type server struct {
	addr, port string
}

// newServer creates a rpc server object.
func newServer(addr, port string) *server {
	return &server{
		addr: addr,
		port: port,
	}
}

// start starts to listen and serve RPCs.
func (s *server) start(c chan error) {
	ln, err := net.Listen("tcp", net.JoinHostPort(s.addr, s.port))
	if err != nil {
		c <- err
		return
	}

	g := grpc.NewServer()
	// GracefulStop stops the server to accept new connections and RPCs
	// and blocks until all the pending RPCs are finished.
	defer g.GracefulStop()

	mdspb.RegisterMdsServer(g, s)
	if err = g.Serve(ln); err != nil {
		c <- err
	}
}

func (s *server) T(ctx context.Context, in *swim.Test) (*swim.Test, error) {
	return nil, nil
}
