package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/kv"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server handles clients requests.
type Server struct {
	cfg *config.Gw

	nilMux   *nilmux.NilMux
	nilLayer *nilmux.Layer

	httpMux   *mux.Router
	httpLayer *nilmux.Layer
	httpSrv   *http.Server

	authCache kv.DB
}

// New creates a server object.
func New(cfg *config.Gw) (*Server, error) {
	log = mlog.GetLogger()

	// Resolve gateway addres.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		cfg:       cfg,
		authCache: kv.New(),
	}

	// Create a rpc layer.
	rpcTypeBytes := []byte{
		0x01, // rpcRaft
		0x02, // rpcNil
	}
	srv.nilLayer = nilmux.NewLayer(rpcTypeBytes, resolvedAddr, true)

	// Create a http layer.
	httpBytes := []byte{
		0x44, // 'D' of DELETE
		0x47, // 'G' of GET
		0x50, // 'P' of POST, PUT
	}
	srv.httpLayer = nilmux.NewLayer(httpBytes, resolvedAddr, true)

	// Create a mux and register layers.
	srv.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	srv.nilMux.RegisterLayer(srv.nilLayer)
	srv.nilMux.RegisterLayer(srv.httpLayer)

	// Create a http server.
	srv.httpMux = mux.NewRouter()
	srv.registerS3Handler(srv.httpMux)
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
	go s.nilMux.ListenAndServeTLS()
	go s.serveNil(s.nilLayer)
	go s.httpSrv.Serve(s.httpLayer)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			log.Info("Received stop signal from OS")
			return s.stop()
		}
	}
}

// stop cleans up the services and shut down the server.
func (s *Server) stop() error {
	// nilMux will closes listener and all the registered layers.
	if err := s.nilMux.Close(); err != nil {
		return err
	}

	// Close the http server.
	return s.httpSrv.Close()
}
