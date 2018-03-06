package swim

import (
	"fmt"
	"log"
	"net/rpc"
	"runtime"
	"sync/atomic"
	"time"
)

// Server maintains an list of connected peers and handles
// gossip messages which incoming periodically. It generates
// and send gossip message periodically and disseminates the
// status of the member if the status is changed.
type Server struct {
	conf Config
	meml *memList

	trans      Transport
	rpcSrv     *rpc.Server
	RPCHandler RPCHandler

	c       chan PingError
	stop    chan chan error
	stopped uint32
}

// NewServer creates swim server object.
func NewServer(conf *Config, trans Transport) (*Server, error) {
	if err := validateConfig(conf); err != nil {
		return nil, err
	}

	memList := newMemList(conf.ID)

	// Make member myself and add to the list.
	me := Member{
		ID:          conf.ID,
		Address:     conf.Address,
		Type:        conf.Type,
		Status:      Alive,
		Incarnation: 0,
	}
	memList.set(me)

	s := &Server{
		conf:    *conf,
		meml:    memList,
		trans:   trans,
		rpcSrv:  rpc.NewServer(),
		stop:    make(chan chan error, 1),
		stopped: uint32(1),
	}
	if err := s.registerRPCHandler(); err != nil {
		return nil, err
	}
	if err := s.rpcSrv.RegisterName(rpcPrefix, s.RPCHandler); err != nil {
		return nil, err
	}

	return s, nil
}

// Serve starts gossiping.
func (s *Server) Serve(c chan PingError) {
	if s.canStart() == false {
		c <- PingError{Err: ErrRunning}
		return
	}

	// Set channel to notify errors to caller.
	s.c = c

	go s.serve()
	runtime.Gosched()

	// Try to join the membership.
	// If failed, sends error message thru channel and stop serving.
	if err := s.join(); err != nil {
		s.c <- PingError{Err: err}
		return
	}

	// Receive ping error thru this channel.
	pec := make(chan PingError, 1)

	// ticker gives signal periodically to send a ping.
	ticker := time.NewTicker(s.conf.PingPeriod)
	for {
		select {
		case exit := <-s.stop:
			// Leaving from the membership.
			// Send good-bye to all members.
			s.leave()
			exit <- nil
			return
		case <-ticker.C:
			go s.ping(pec)
		case pe := <-pec:
			go s.handleErr(pe, pec)
			s.c <- pe
		}
	}
}

// Stop will stop the swim server and cleaning up.
func (s *Server) Stop() error {
	if s.isStopped() {
		return ErrStopped
	}

	exit := make(chan error)
	s.stop <- exit

	atomic.SwapUint32(&s.stopped, uint32(1))

	return <-exit
}

// GetMap returns cluster map.
func (s *Server) GetMap() []Member {
	return s.meml.fetch(0)
}

// GetMDS returns an address of the MDS server.
func (s *Server) GetMDS() (string, error) {
	mems := s.meml.fetch(0)
	for _, m := range mems {
		if m.Type == MDS && m.Status == Alive {
			return string(m.Address), nil
		}
	}

	return "", fmt.Errorf("no alive mds in this cluster")
}

func (s *Server) registerRPCHandler() (err error) {
	s.RPCHandler, err = newRPCHandler(s)
	return
}

func (s *Server) canStart() bool {
	return atomic.SwapUint32(&s.stopped, uint32(0)) == 1
}

func (s *Server) isStopped() bool {
	return atomic.LoadUint32(&s.stopped) == 1
}

func (s *Server) serve() {
	for {
		conn, err := s.trans.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go s.rpcSrv.ServeConn(conn)
	}
}
