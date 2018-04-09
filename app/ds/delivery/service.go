package delivery

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/ds/usecase/admin"
	"github.com/chanyoung/nil/app/ds/usecase/membership"
	"github.com/chanyoung/nil/app/ds/usecase/object"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type Service struct {
	nilMux *nilmux.NilMux

	adminL      *nilmux.Layer
	objectL     *nilmux.Layer
	membershipL *nilmux.Layer

	httpHandler http.Handler
	httpSrv     *http.Server

	adminSrv      *rpc.Server
	adminHandlers admin.AdminHandlers

	membershipHandler membership.MembershipHandlers
}

// NewDeliveryService creates a delivery service with necessary dependencies.
func NewDeliveryService(cfg *config.Ds, ah admin.AdminHandlers, oh object.ObjectHandlers, mh membership.MembershipHandlers) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid nil arguments")
	}
	logger = mlog.GetPackageLogger("app/ds/delivery")

	// Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "resolve gateway address failed")
	}

	// Create transport layers.
	adminL := nilmux.NewLayer(adminTypeBytes(), rAddr, false)
	objectL := nilmux.NewLayer(objectTypeBytes(), rAddr, true)
	membershipL := nilmux.NewLayer(membershipTypeBytes(), rAddr, false)

	// Create a mux and register layers.
	m := nilmux.NewNilMux(addr, &cfg.Security)
	m.RegisterLayer(adminL)
	m.RegisterLayer(objectL)
	m.RegisterLayer(membershipL)

	// Create a http handler.
	h := makeHandler(oh)

	// Create http server.
	hsrv := &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.New(logger.Writer(), "http server", log.Lshortfile),
	}

	// Create swim server.
	if err := mh.Create(membershipL); err != nil {
		return nil, err
	}

	// Create admin server.
	ads := rpc.NewServer()
	if err := ads.RegisterName(nilrpc.DSRPCPrefix, ah); err != nil {
		return nil, err
	}

	return &Service{
		nilMux: m,

		adminL:  adminL,
		objectL: objectL,

		httpHandler: h,
		httpSrv:     hsrv,

		membershipHandler: mh,

		adminSrv:      ads,
		adminHandlers: ah,
	}, nil
}

// Run starts the gateway delivery service.
func (s *Service) Run() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.Run")
	ctxLogger.Info("Start gateway delivery service ...")

	go s.nilMux.ListenAndServeTLS()
	go s.serveAdmin()
	go s.httpSrv.Serve(s.objectL)
	go s.membershipHandler.Run()
}

// Stop cleans up the services and shut down the server.
func (s *Service) Stop() error {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.Stop")
	ctxLogger.Info("Stop gateway delivery service ...")

	// nilMux will closes listener and all the registered layers.
	if err := s.nilMux.Close(); err != nil {
		return errors.Wrap(err, "close nil mux failed")
	}

	// Close the http server.
	return s.httpSrv.Close()
}

func (s *Service) serveAdmin() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.serveAdmin")

	for {
		conn, err := s.adminL.Accept()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "accept connection from admin layer failed"))
			return
		}
		go s.adminSrv.ServeConn(conn)
	}
}
