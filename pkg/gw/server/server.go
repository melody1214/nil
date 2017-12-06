package server

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/gw/mdsmap"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Server handles clients requests.
type Server struct {
	cfg    *config.Gw
	mdsMap *mdsmap.MdsMap
	srv    *http.Server
}

// New creates a server object.
func New(cfg *config.Gw) (*Server, error) {
	log = mlog.GetLogger()

	// Make mds map.
	mm, err := mdsmap.New(&cfg.Security)
	if err != nil {
		return nil, err
	}

	// Get TLS config.
	var tlsCfg *tls.Config
	if cfg.UseHTTPS == "true" {
		tlsCfg = security.DefaultTLSConfig()
	}

	srv := &Server{
		cfg:    cfg,
		mdsMap: mm,

		srv: &http.Server{
			Addr:           net.JoinHostPort(cfg.ServerAddr, cfg.ServerPort),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
			TLSConfig:      tlsCfg,
		},
	}
	srv.initHandler()

	return srv, nil
}

// Start starts to listen and serve requests.
func (s *Server) Start() error {
	// Filling mds map information.
	if err := s.mdsMap.Start(s.cfg.FirstMds); err != nil {
		return err
	}

	// Http server runs and return error through the httpc channel.
	httpc := make(chan error)
	go func() {
		httpc <- s.srv.ListenAndServeTLS(
			s.cfg.Security.CertsDir+"/"+s.cfg.Security.ServerCrt,
			s.cfg.Security.CertsDir+"/"+s.cfg.Security.ServerKey,
		)
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
