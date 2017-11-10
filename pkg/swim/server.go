package swim

import (
	"sync/atomic"
	"time"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

// Server has functions
type Server struct {
	id      string
	meml    *memList
	stop    chan chan error
	stopped uint32
}

// NewServer creates swim server object.
func NewServer(id string, addr, port string) *Server {
	// Make member myself and add to the list.
	me := newMember(id, addr, port, swimpb.Status_ALIVE, 0)
	memList := newMemList()

	memList.set(me)

	return &Server{
		id:      id,
		meml:    memList,
		stop:    make(chan chan error, 1),
		stopped: uint32(1),
	}
}

// Serve starts gossiping.
func (s *Server) Serve(c chan PingError) {
	if s.canStart() == false {
		c <- PingError{Err: ErrRunning}
		return
	}

	// Try to join the membership.
	// If failed, sends error message thru channel and stop serving.
	if err := s.join(); err != nil {
		c <- PingError{Err: err}
		return
	}

	// Receive ping error thru this channel.
	pec := make(chan PingError, 1)

	// ticker gives signal periodically to send a ping.
	ticker := time.NewTicker(5 * time.Second)
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
			c <- pe
		}
	}
}

func (s *Server) canStart() bool {
	return atomic.SwapUint32(&s.stopped, uint32(0)) == 1
}

func (s *Server) isStopped() bool {
	return atomic.LoadUint32(&s.stopped) == 1
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
func (s *Server) GetMap() []*swimpb.Member {
	return s.meml.fetch(0)
}
