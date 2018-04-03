package delivery

import (
	"net"
	"net/http"
	"time"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type Service struct {
	ah AdminHandlers
	ch ClientHandlers

	nilMux *nilmux.NilMux

	rpcL  *nilmux.Layer
	httpL *nilmux.Layer

	httpHandler http.Handler
	httpSrv     *http.Server
}

func NewDeliveryService(cfg *config.Gw, ah AdminHandlers, ch ClientHandlers) (*Service, error) {
	l := mlog.GetLogger()
	if l == nil {
		return nil, errors.New("failed to get logger")
	}
	log = l.WithField("package", "delivery")

	addr := cfg.ServerAddr + ":" + cfg.ServerPort

	if cfg == nil || ah == nil || ch == nil {
		return nil, errors.New("invalid nil arguments")
	}

	// 1. Resolve gateway address.
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "resolve gateway address failed")
	}

	// 2. Create transport layers.
	rpcL := nilmux.NewLayer(rpcTypeBytes(), rAddr, true)
	httpL := nilmux.NewLayer(httpTypeBytes(), rAddr, true)

	// 3. Create a mux and register layers.
	m := nilmux.NewNilMux(addr, &cfg.Security)
	m.RegisterLayer(rpcL)
	m.RegisterLayer(httpL)

	// 4. Create a http handler.
	h := makeHandler(ch)

	// 5. Create http server.
	hsrv := &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &Service{
		ah: ah,
		ch: ch,

		nilMux: m,

		rpcL:  rpcL,
		httpL: httpL,

		httpHandler: h,
		httpSrv:     hsrv,
	}, nil
}

// Run starts the gateway delivery service.
func (s *Service) Run() {
	ctxLogger := log.WithField("method", "Service.Run")
	ctxLogger.Info("Start gateway delivery service ...")

	go s.nilMux.ListenAndServeTLS()
	go s.handleAdmin()
	go s.httpSrv.Serve(s.httpL)
}

// Stop cleans up the services and shut down the server.
func (s *Service) Stop() error {
	// nilMux will closes listener and all the registered layers.
	if err := s.nilMux.Close(); err != nil {
		return errors.Wrap(err, "close nil mux failed")
	}

	// Close the http server.
	return s.httpSrv.Close()
}

func (s *Service) handleAdmin() {
	ctxLogger := log.WithField("method", "Service.handleAdmin")

	for {
		conn, err := s.rpcL.Accept()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "accept connection from nil layer failed"))
			return
		}

		go s.ah.Proxying(conn)
	}
}

type AdminHandlers interface {
	Proxying(conn net.Conn)
}

type ClientHandlers interface {
	MakeBucketHandler(w http.ResponseWriter, r *http.Request)
	RemoveBucketHandler(w http.ResponseWriter, r *http.Request)

	PutObjectHandler(w http.ResponseWriter, r *http.Request)
	GetObjectHandler(w http.ResponseWriter, r *http.Request)
	DeleteObjectHandler(w http.ResponseWriter, r *http.Request)
}
