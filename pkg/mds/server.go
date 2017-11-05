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
	cfg *Config
}

// newServer creates a rpc server object.
func newServer(cfg *Config) *server {
	return &server{
		cfg: cfg,
	}
}

// start starts to listen and serve RPCs.
func (s *server) start(c chan error) {
	ln, err := net.Listen("tcp", net.JoinHostPort(s.cfg.ServerAddr, s.cfg.ServerPort))
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
