package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server handles clients requests.
type Server struct {
	cfg *config.Gw

	httpSrv *http.Server
	httpMux *mux.Router
	httpTr  *httpTransportLayer
}

// New creates a server object.
func New(cfg *config.Gw) (*Server, error) {
	log = mlog.GetLogger()

	// Resolve gateway addres.
	addr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		cfg:     cfg,
		httpTr:  newHTTPTransportLayer(addr),
		httpMux: mux.NewRouter(),
	}

	srv.registerHTTPHandler()

	srv.httpSrv = &http.Server{
		Handler:        srv.httpMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return srv, nil
}

// Start starts to listen and serve requests.
func (s *Server) Start() error {
	go s.listenAndServeTLS()
	go s.httpSrv.Serve(s.httpTr)

	// Http server runs and return error through the httpc channel.
	httpc := make(chan error)

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
