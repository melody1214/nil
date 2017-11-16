package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/pkg/gw/s3"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server handles clients requests.
type Server struct {
	cfg *config.Gw

	mux *mux.Router
}

// New creates a server object.
func New(cfg *config.Gw) *Server {
	log = mlog.GetLogger()

	return &Server{
		cfg: cfg,
		mux: mux.NewRouter(),
	}
}

// Start starts to listen and serve requests.
func (s *Server) Start() error {
	// Register s3 API router.
	if err := s3.RegisterS3APIRouter(s.cfg, s.mux); err != nil {
		return err
	}

	// Http server runs and return error through the httpc channel.
	httpc := make(chan error)
	go func() {
		httpc <- http.ListenAndServe(net.JoinHostPort(s.cfg.ServerAddr, s.cfg.ServerPort), s.mux)
	}()

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			log.Info("Received stop signal from OS")
			return s.stop()
		case err := <-httpc:
			log.Error(err)
			return s.stop()
		}
	}
}

// stop cleans up the services and shut down the server.
func (s *Server) stop() error {
	return nil
}
