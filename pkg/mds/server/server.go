package server

import (
	"net"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"github.com/chanyoung/nil/pkg/util/config"
	"google.golang.org/grpc"
)

// Server serve RPCs.
type Server struct {
	cfg  *config.Mds
	swim *swim.Server
}

// New creates a rpc server object.
func New(cfg *config.Mds) *Server {
	return &Server{
		cfg:  cfg,
		swim: swim.NewServer(cfg.ID, cfg.ServerAddr, cfg.ServerPort),
	}
}

// Start starts to listen and serve RPCs.
func (s *Server) Start(c chan error) {
	// Try to grab a free port which will serve gRPC.
	ln, err := net.Listen("tcp", net.JoinHostPort(s.cfg.ServerAddr, s.cfg.ServerPort))
	if err != nil {
		c <- err
		return
	}

	g := grpc.NewServer()
	// GracefulStop stops the server to accept new connections and RPCs
	// and blocks until all the pending RPCs are finished.
	defer g.GracefulStop()

	// Register swim gossip protocol service to grpc server.
	swimpb.RegisterSwimServer(g, s.swim)

	// Register MDS service to grpc server.
	mdspb.RegisterMdsServer(g, s)

	if err = g.Serve(ln); err != nil {
		c <- err
	}
}
