package server

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var log *logrus.Logger

// Server serve RPCs.
type Server struct {
	cfg  *config.Mds
	g    *grpc.Server
	swim *swim.Server
}

// New creates a rpc server object.
func New(cfg *config.Mds) *Server {
	log = mlog.GetLogger()

	return &Server{
		cfg:  cfg,
		swim: swim.NewServer(cfg.ID, cfg.ServerAddr, cfg.ServerPort),
	}
}

// Start starts to listen and serve RPCs.
func (s *Server) Start() error {
	// Try to grab a free port which will serve gRPC.
	ln, err := net.Listen("tcp", net.JoinHostPort(s.cfg.ServerAddr, s.cfg.ServerPort))
	if err != nil {
		return err
	}

	// Creates new grpc server for serving MDS requests.
	s.g = grpc.NewServer()

	// Register swim gossip protocol service to grpc server.
	swimpb.RegisterSwimServer(s.g, s.swim)

	// Register MDS service to grpc server.
	mdspb.RegisterMdsServer(s.g, s)

	// Starts gRPC service.
	gc := make(chan error, 1)
	go func() {
		if err = s.g.Serve(ln); err != nil {
			gc <- err
		}
	}()

	// Starts swim service.
	sc := make(chan error, 1)
	go s.swim.Serve(sc)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case err := <-gc:
			log.Error(err)
		case err := <-sc:
			log.Error(err)
		case <-sigc:
			log.Info("Received stop signal from OS")
			return s.stop()
		}
	}
}

// stop cleans up the services and shut down the server.
func (s *Server) stop() error {
	// Stop swim server and leaving from the membership.
	s.swim.Stop()

	// GracefulStop stops the server to accept new connections and RPCs
	// and blocks until all the pending RPCs are finished.
	s.g.GracefulStop()

	return nil
}