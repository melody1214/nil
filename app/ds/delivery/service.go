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
	"github.com/chanyoung/nil/pkg/cluster"
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
	adminHandlers admin.Handlers

	cs                *cluster.Service
	membershipHandler membership.Handlers
}

// SetupDeliveryService creates a delivery service with necessary dependencies.
func SetupDeliveryService(cfg *config.Ds, ah admin.Handlers, oh object.Handlers, cs *cluster.Service) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid nil arguments")
	}
	logger = mlog.GetPackageLogger("app/ds/delivery")

	s := &Service{
		adminHandlers: ah,
		cs:            cs,
	}

	// Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "resolve gateway address failed")
	}

	// Create transport layers.
	s.adminL = nilmux.NewLayer(adminTypeBytes(), rAddr, false)
	s.objectL = nilmux.NewLayer(objectTypeBytes(), rAddr, true)
	s.membershipL = nilmux.NewLayer(membershipTypeBytes(), rAddr, false)

	// Create a mux and register layers.
	s.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	s.nilMux.RegisterLayer(s.adminL)
	s.nilMux.RegisterLayer(s.objectL)
	s.nilMux.RegisterLayer(s.membershipL)

	// Create a http handler.
	s.httpHandler = makeHandler(oh)

	// Create http server.
	s.httpSrv = &http.Server{
		Handler:        s.httpHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.New(logger.Writer(), "http server", log.Lshortfile),
	}

	// // Create swim server.
	// if err := mh.Create(membershipL); err != nil {
	// 	return nil, err
	// }

	// Create admin server.
	s.adminSrv = rpc.NewServer()
	if err := s.adminSrv.RegisterName(nilrpc.DSRPCPrefix, s.adminHandlers); err != nil {
		return nil, err
	}

	// Run the delivery server.
	s.run()

	// Setup the membership server and run.
	clusterConf := cluster.DefaultConfig()
	clusterConf.Name = cluster.NodeName(cfg.ID)
	clusterConf.Address = cluster.NodeAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	clusterConf.Coordinator = cluster.NodeAddress(cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(cfg.Swim.Period); err == nil {
		clusterConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(cfg.Swim.Expire); err == nil {
		clusterConf.PingExpire = t
	}
	clusterConf.Type = cluster.DS
	if err := s.cs.StartMembershipServer(*clusterConf, nilmux.NewSwimTransportLayer(s.membershipL)); err != nil {
		return nil, err
	}
	// Join the local cluster.
	conn, err := nilrpc.Dial(clusterConf.Coordinator.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &nilrpc.MCLJoinRequest{
		Node: cluster.Node{
			Name: clusterConf.Name,
			Type: clusterConf.Type,
			Stat: cluster.Alive,
			Addr: clusterConf.Address,
		},
	}
	res := &nilrpc.MCLJoinResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClustermapJoin.String(), req, res); err != nil {
		return nil, err
	}

	return s, nil

	// return &Service{
	// 	nilMux: m,

	// 	adminL:      adminL,
	// 	objectL:     objectL,
	// 	membershipL: membershipL,

	// 	httpHandler: h,
	// 	httpSrv:     hsrv,

	// 	membershipHandler: mh,

	// 	adminSrv:      ads,
	// 	adminHandlers: ah,
	// }, nil
}

// Run starts the gateway delivery service.
func (s *Service) run() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.Run")
	ctxLogger.Info("Start gateway delivery service ...")

	go s.nilMux.ListenAndServeTLS()
	go s.serveAdmin()
	go s.httpSrv.Serve(s.objectL)
	// go s.membershipHandler.Run(s.membershipL)
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
