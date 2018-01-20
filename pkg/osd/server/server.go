package server

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server serve user requests and management orders.
type Server struct {
	cfg *config.Osd

	nilMux        *nilmux.NilMux
	nilLayer      *nilmux.Layer
	nilRPCSrv     *rpc.Server
	NilRPCHandler NilRPCHandler

	swimTransportLayer *nilmux.SwimTransportLayer
	swimLayer          *nilmux.Layer
	swimSrv            *swim.Server
}

// New creates a server object.
func New(cfg *config.Osd) (*Server, error) {
	log = mlog.GetLogger()

	// Resolve gateway addres.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		cfg: cfg,
	}

	// Create a rpc layer.
	rpcTypeBytes := []byte{
		0x02, // rpcNil
	}
	srv.nilLayer = nilmux.NewLayer(rpcTypeBytes, resolvedAddr, false)

	swimTypeBytes := []byte{
		0x03, // rpcSwim
	}
	srv.swimLayer = nilmux.NewLayer(swimTypeBytes, resolvedAddr, false)
	srv.swimTransportLayer = nilmux.NewSwimTransportLayer(srv.swimLayer)

	swimConf := swim.DefaultConfig()
	swimConf.ID = swim.ServerID(cfg.ID)
	swimConf.Address = swim.ServerAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	swimConf.Type = swim.OSD

	srv.swimSrv, err = swim.NewServer(
		swimConf,
		srv.swimTransportLayer,
	)
	if err != nil {
		return nil, err
	}

	// Create a mux and register layers.
	srv.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	srv.nilMux.RegisterLayer(srv.nilLayer)
	srv.nilMux.RegisterLayer(srv.swimLayer)

	// Create nil RPC server.
	srv.nilRPCSrv = rpc.NewServer()
	srv.newNilRPCHandler()
	if err := srv.nilRPCSrv.Register(srv.NilRPCHandler); err != nil {
		return nil, err
	}

	return srv, nil
}

// Start starts to listen and serve RPCs.
func (s *Server) Start() error {
	// Start tcp listen and serve.
	go s.nilMux.ListenAndServeTLS()
	go s.serveNilRPC(s.nilLayer)

	// Starts swim service.
	sc := make(chan swim.PingError, 1)
	go s.swimSrv.Serve(sc)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case err := <-sc:
			log.WithFields(logrus.Fields{
				"server":       "swim",
				"message type": err.Type,
				"destID":       err.DestID,
			}).Error(err.Err)
		case <-sigc:
			log.Info("Received stop signal from OS")
			return s.stop()
		}
	}
}

// stop cleans up the services and shut down the server.
func (s *Server) stop() error {
	return nil
}
