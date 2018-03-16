package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/gw/server/rpchandling"
	"github.com/chanyoung/nil/pkg/gw/server/s3handling"
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

	rpcHandler *rpchandling.Handler
	s3Handler  *s3handling.Handler

	nilMux   *nilmux.NilMux
	nilLayer *nilmux.Layer

	httpMux   *mux.Router
	httpLayer *nilmux.Layer
	httpSrv   *http.Server
}

// New creates a server object.
func New(cfg *config.Gw) (*Server, error) {
	// 1. Get logger.
	log = mlog.GetLogger()

	// 2. Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
	}

	// 3. Create server object with the config.
	srv := &Server{cfg: cfg}

	// 4. Create each handlers.
	srv.rpcHandler = rpchandling.NewHandler()
	srv.s3Handler = s3handling.NewHandler()

	// 5. Create each handling layers.
	srv.nilLayer = nilmux.NewLayer(rpchandling.TypeBytes(), resolvedAddr, true)
	srv.httpLayer = nilmux.NewLayer(s3handling.TypeBytes(), resolvedAddr, true)

	// 6. Create a mux and register handling layers.
	srv.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	srv.nilMux.RegisterLayer(srv.nilLayer)
	srv.nilMux.RegisterLayer(srv.httpLayer)

	// 7. Create a http multiplexer.
	srv.httpMux = mux.NewRouter()

	// 8. Register s3 handling layer.
	srv.s3Handler.Register(srv.httpMux)

	// 9. Create http server with the given mux and settings.
	srv.httpSrv = &http.Server{
		Handler:        srv.httpMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 10. Prepare cluster map.
	go prepareClusterMap(cfg.FirstMds)

	return srv, nil
}

// Start starts to listen and serve requests.
func (s *Server) Start() error {
	go s.nilMux.ListenAndServeTLS()
	go s.serveRPC()
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

func (s *Server) serveRPC() {
	for {
		conn, err := s.nilLayer.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		go s.rpcHandler.Proxying(conn)
	}
}

// prepareClusterMap prepares the cluster map with the given address.
// It calls only one time in the initiating routine and retry repeatedly
// until it succeed.
func prepareClusterMap(mdsAddr string) {
	for {
		time.Sleep(100 * time.Millisecond)

		m, err := cmap.GetLatest(mdsAddr)
		if err != nil {
			log.Error(err)
			continue
		}

		_, err = m.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			log.Error(err)
			continue
		}

		break
	}
}
