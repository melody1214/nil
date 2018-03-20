package gw

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/gw/rpchandling"
	"github.com/chanyoung/nil/pkg/gw/s3handling"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

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
		return nil, err
	}
	log = mlog.GetLogger()
	if log == nil {
		return nil, errors.New("nil logger object")
	}
	log.WithField("location", cfg.LogLocation).Info("Setting logger succeeded")

	// 2. Generate gateway ID.
	cfg.ID = uuid.Gen()
	log.WithField("uuid", cfg.ID).Info("Generating gateway UUID succeeded")

	// 3. Creates gateway server object with the given config.
	g := &Gw{cfg: cfg}

	// 4. Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	resolvedAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, err
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

	// 11. Prepare cluster map.
	go prepareClusterMap(cfg.FirstMds)

	log.Info("Creating gateway object succeeded")

	return g, nil
}

// Start starts the gateway.
func (g *Gw) Start() {
	log.Info("Start gateway service ...")

	go g.nilMux.ListenAndServeTLS()
	go g.serveRPC()
	go g.httpSrv.Serve(g.httpLayer)

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			log.Info("Received stop signal from OS")
			g.stop()
			return
		}
	}
}

// stop cleans up the services and shut down the server.
func (g *Gw) stop() error {
	// nilMux will closes listener and all the registered layers.
	if err := g.nilMux.Close(); err != nil {
		return err
	}

	// Close the http server.
	return g.httpSrv.Close()
}

func (g *Gw) serveRPC() {
	for {
		conn, err := g.nilLayer.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		go g.rpcHandler.Proxying(conn)
	}
}

// prepareClusterMap prepares the cluster map with the given address.
// It calls only one time in the initiating routine and retry repeatedly
// until it succeed.
func prepareClusterMap(mdsAddr string) {
	for {
		time.Sleep(100 * time.Millisecond)

		m, err := cmap.GetLatest(cmap.FromRemote(mdsAddr))
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
