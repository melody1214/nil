package delivery

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/ds/usecase/cluster"
	"github.com/chanyoung/nil/app/ds/usecase/gencoding"
	"github.com/chanyoung/nil/app/ds/usecase/object"
	"github.com/chanyoung/nil/pkg/cmap"
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

	rpcL        *nilmux.Layer
	httpL       *nilmux.Layer
	membershipL *nilmux.Layer

	httpHandler http.Handler
	httpSrv     *http.Server

	rpcSrv *rpc.Server

	cls cluster.Service
	obh object.Handlers
	cms *cmap.Service
	ges gencoding.Service
}

// SetupDeliveryService creates a delivery service with necessary dependencies.
func SetupDeliveryService(cfg *config.Ds, cls cluster.Service, obh object.Handlers, cms *cmap.Service, ges gencoding.Service) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid nil arguments")
	}
	logger = mlog.GetPackageLogger("app/ds/delivery")
	ctxLogger := mlog.GetFunctionLogger(logger, "SetupDeliveryService")

	s := &Service{
		cls: cls,
		obh: obh,
		cms: cms,
		ges: ges,
	}

	// Resolve gateway address.
	addr := cfg.ServerAddr + ":" + cfg.ServerPort
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "resolve gateway address failed")
	}

	// Create transport layers.
	s.rpcL = nilmux.NewLayer(rpcTypeBytes(), rAddr, false)
	s.httpL = nilmux.NewLayer(httpTypeBytes(), rAddr, true)
	s.membershipL = nilmux.NewLayer(membershipTypeBytes(), rAddr, false)

	// Create a mux and register layers.
	s.nilMux = nilmux.NewNilMux(addr, &cfg.Security)
	s.nilMux.RegisterLayer(s.rpcL)
	s.nilMux.RegisterLayer(s.httpL)
	s.nilMux.RegisterLayer(s.membershipL)

	// Create a http handler.
	s.httpHandler = makeHandler(obh)

	// Create http server.
	s.httpSrv = &http.Server{
		Handler:        s.httpHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.New(logger.Writer(), "http server", log.Lshortfile),
	}

	// Create rpc server.
	s.rpcSrv = rpc.NewServer()
	if err := s.rpcSrv.RegisterName(nilrpc.DsClusterPrefix, s.cls); err != nil {
		return nil, err
	}
	if err := s.rpcSrv.RegisterName(nilrpc.DsGencodingPrefix, s.ges); err != nil {
		return nil, err
	}
	if err := s.rpcSrv.RegisterName(nilrpc.DsObjectPrefix, s.obh.RPCHandler()); err != nil {
		return nil, err
	}

	// Run the delivery server.
	if err := s.nilMux.ListenAndServeTLS(); err != nil {
		return nil, errors.Wrap(err, "failed to listen and serve TLS")
	}
	ctxLogger.Info("start to listen main mux")
	go s.serveRPC()
	ctxLogger.Info("start to serve rpc requests")
	go s.serveHTTP()
	ctxLogger.Info("start to serve http requests")

	// Setup the membership server and run.
	cmapConf := cmap.DefaultConfig()
	cmapConf.Name = cmap.NodeName(cfg.ID)
	cmapConf.Address = cmap.NodeAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	cmapConf.Coordinator = cmap.NodeAddress(cfg.Swim.CoordinatorAddr)
	if t, err := time.ParseDuration(cfg.Swim.Period); err == nil {
		cmapConf.PingPeriod = t
	}
	if t, err := time.ParseDuration(cfg.Swim.Expire); err == nil {
		cmapConf.PingExpire = t
	}
	cmapConf.Type = cmap.DS
	if err := s.cms.StartMembershipServer(*cmapConf, nilmux.NewSwimTransportLayer(s.membershipL)); err != nil {
		return nil, err
	}
	// Join the local cmap.
	conn, err := nilrpc.Dial(cmapConf.Coordinator.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &nilrpc.MCLLocalJoinRequest{
		Node: cmap.Node{
			Name: cmapConf.Name,
			Type: cmapConf.Type,
			Stat: cmap.NodeAlive,
			Addr: cmapConf.Address,
		},
	}
	res := &nilrpc.MCLLocalJoinResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClusterLocalJoin.String(), req, res); err != nil {
		return nil, err
	}

	return s, nil
}

// Serve http server.
func (s *Service) serveHTTP() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.serveHTTP")

	for {
		if err := s.httpSrv.Serve(s.httpL); err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to serve http"))
			ctxLogger.Error("please inspect this. retry to serve after 10 sec.")
			time.Sleep(10 * time.Second)
		}
	}
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

func (s *Service) serveRPC() {
	ctxLogger := mlog.GetMethodLogger(logger, "Service.serveAdmin")

	for {
		conn, err := s.rpcL.Accept()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "accept connection from admin layer failed"))
			return
		}
		go s.rpcSrv.ServeConn(conn)
	}
}
