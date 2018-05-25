package cmap

import (
	"fmt"
	"net"
	"net/rpc"
	"runtime"
	"sync/atomic"
	"time"
)

// Server maintains an list of connected peers and handles
// gossip messages which incoming periodically. It generates
// and send gossip message periodically and disseminates the
// status of the member if the status is changed.
type server struct {
	cfg     Config
	manager *manager

	trans      Transport
	rpcSrv     *rpc.Server
	rpcHandler rpcHandler

	stop    chan chan error
	stopped uint32
}

// newServer creates swim server object.
func newServer(cfg Config, manager *manager, trans Transport) (*server, error) {
	s := &server{
		cfg:     cfg,
		manager: manager,
		trans:   trans,
		rpcSrv:  rpc.NewServer(),
		stop:    make(chan chan error, 1),
		stopped: uint32(1),
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
		case exit := <-s.stop:
			// Leaving from the membership.
			// Send good-bye to all members.
			s.leave()
			exit <- nil
			return
		case <-ticker.C:
			go s.ping()
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

// Halt will stop the swim server and cleaning up.
func (s *server) halt() error {
	if s.isStopped() {
		return fmt.Errorf("server is already stopped")
	}

	exit := make(chan error)
	s.stop <- exit

	atomic.SwapUint32(&s.stopped, uint32(1))

	return <-exit
}

func (s *server) canStart() bool {
	return atomic.SwapUint32(&s.stopped, uint32(0)) == 1
}

func (s *server) isStopped() bool {
	return atomic.LoadUint32(&s.stopped) == 1
}

// Transport is swim network transport abstraction layer.
type Transport interface {
	net.Listener

	Dial(address string, timeout time.Duration) (net.Conn, error)
}
