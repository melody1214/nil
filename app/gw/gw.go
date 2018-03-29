package gw

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/app/gw/rpchandling"
	"github.com/chanyoung/nil/app/gw/s3handling"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// Gw is the [project name] gateway server node.
type Gw struct {
	// cfg is pointing the gateway config from command package.
	cfg *config.Gw

	rpcHandler *rpchandling.Handler
	s3Handler  *s3handling.Handler

	nilMux   *nilmux.NilMux
	nilLayer *nilmux.Layer

	httpMux   *mux.Router
	httpLayer *nilmux.Layer
	httpSrv   *http.Server
}

// New creates a gateway object.
func New(cfg *config.Gw) (*Gw, error) {
	// 1. Setting logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return nil, errors.Wrap(err, "init log failed")
	}
	log = mlog.GetLogger().WithField("package", "gw")
	if log == nil {
		return nil, errors.New("init log failed: nil logger object")
	}
	ctxLogger := log.WithField("method", "New")
	ctxLogger.Info("Setting logger succeeded")

	// 2. Generate gateway ID.
	cfg.ID = uuid.Gen()
	ctxLogger.WithField("uuid", cfg.ID).Info("Generating gateway UUID succeeded")

	// 3. Creates gateway server object with the given config.
	g := &Gw{cfg: cfg}

	// 4. Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, errors.Wrap(err, "resolve gateway address failed")
	}

	// 5. Create each handlers.
	g.rpcHandler = rpchandling.NewHandler()
	g.s3Handler = s3handling.NewHandler()

	// 6. Create each handling layers.
	g.nilLayer = nilmux.NewLayer(rpchandling.TypeBytes(), resolvedAddr, true)
	g.httpLayer = nilmux.NewLayer(s3handling.TypeBytes(), resolvedAddr, true)

	// 7. Create a mux and register handling layers.
	g.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	g.nilMux.RegisterLayer(g.nilLayer)
	g.nilMux.RegisterLayer(g.httpLayer)

	// 8. Create a http multiplexer.
	g.httpMux = mux.NewRouter()

	// 9. Register s3 handling layer.
	g.s3Handler.RegisteredTo(g.httpMux)

	// 10. Create http server with the given mux and settings.
	g.httpSrv = &http.Server{
		Handler:        g.httpMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 11. Prepare the initial cluster map.
	if err := cmap.Initial(cfg.FirstMds); err != nil {
		return nil, errors.Wrap(err, "init cluster map failed")
	}

	ctxLogger.Info("Creating gateway object succeeded")

	return g, nil
}

// Start starts the gateway.
func (g *Gw) Start() {
	ctxLogger := log.WithField("method", "Gw.Start")
	ctxLogger.Info("Start gateway service ...")

	go g.nilMux.ListenAndServeTLS()
	go g.serveRPC()
	go g.httpSrv.Serve(g.httpLayer)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			ctxLogger.Info("Received stop signal from OS")
			g.stop()
			return
		}
	}
}

// stop cleans up the services and shut down the server.
func (g *Gw) stop() error {
	// nilMux will closes listener and all the registered layers.
	if err := g.nilMux.Close(); err != nil {
		return errors.Wrap(err, "close nil mux failed")
	}

	// Close the http server.
	return g.httpSrv.Close()
}

func (g *Gw) serveRPC() {
	for {
		conn, err := g.nilLayer.Accept()
		if err != nil {
			log.WithField("method", "Gw.serveRPC").Error(
				errors.Wrap(
					err,
					"accept connection from nil layer failed",
				),
			)
			return
		}

		go g.rpcHandler.Proxying(conn)
	}
}
