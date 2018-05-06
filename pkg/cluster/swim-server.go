package cluster

import (
	"net"
	"net/rpc"
	"runtime"
	"time"
)

// Server maintains an list of connected peers and handles
// gossip messages which incoming periodically. It generates
// and send gossip message periodically and disseminates the
// status of the member if the status is changed.
type server struct {
	cfg         Config
	cMapManager *cMapManager

	trans      Transport
	rpcSrv     *rpc.Server
	rpcHandler rpcHandler

	// c       chan PingError
	// stop    chan chan error
	// stopped uint32
}

// newServer creates swim server object.
func newServer(cfg Config, cMapManager *cMapManager, trans Transport) (*server, error) {
	s := &server{
		cfg:         cfg,
		cMapManager: cMapManager,
		trans:       trans,
		rpcSrv:      rpc.NewServer(),
		// stop:        make(chan chan error, 1),
		// stopped:     uint32(1),
	}

	s.rpcHandler = s
	if err := s.rpcSrv.RegisterName(rpcPrefix, s.rpcHandler); err != nil {
		return nil, err
	}

	return s, nil
}

// run starts the server.
func (s *server) run() {
	go s.listenAndServe()
	runtime.Gosched()

	// ticker gives signal periodically to send a ping.
	ticker := time.NewTicker(s.cfg.PingPeriod)
	for {
		select {
		// case exit := <-s.stop:
		// 	// Leaving from the membership.
		// 	// Send good-bye to all members.
		// 	s.leave()
		// 	exit <- nil
		// 	return
		case <-ticker.C:
			s.ping()
			// case pe := <-pec:
			// 	go s.handleErr(pe, pec)
			// 	s.c <- pe
		}
	}
}

// listenAndServe listen and serve the swim requests.
func (s *server) listenAndServe() {
	for {
		conn, err := s.trans.Accept()
		if err != nil {
			return
		}
		go s.rpcSrv.ServeConn(conn)
	}
}

// Transport is swim network transport abstraction layer.
type Transport interface {
	net.Listener

	Dial(address string, timeout time.Duration) (net.Conn, error)
}
