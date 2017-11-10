package swim

import (
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"golang.org/x/net/context"
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
func (s *Server) Serve(c chan error) {
	if s.canStart() == false {
		c <- errors.New("swim server is already running")
		return
	}

	// Try to join the membership.
	// If failed, sends error message thru channel and stop serving.
	if err := s.join(); err != nil {
		c <- err
		return
	}

	// ticker gives signal periodically to send a ping.
	ticker := time.NewTicker(5 * time.Second)
	// pending queue for pinging.
	var pending []*swimpb.Member
	for {
		select {
		case exit := <-s.stop:
			// Leaving from the membership.
			// Send good-bye to all members.
			s.leave()
			exit <- nil
			return
		case <-ticker.C:
			// Refill the pending queue.
			if len(pending) == 0 {
				pending = append(pending, s.meml.getAll()...)
			}

			// Fetch first target from pending queue.
			t := pending[0]
			pending = pending[1:]

			// Send ping only the target is not faulty.
			if t.Status == swimpb.Status_FAULTY {
				break
			}

			// Make ping message.
			p := &swimpb.PingMessage{}
			p.Type = swimpb.Type_PING
			p.Memlist = s.meml.getAll()

			// Sends ping message to the target.
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				_, err := s.sendPing(ctx, net.JoinHostPort(t.Addr, t.Port), p)
				if err != nil {
					c <- err
					return
				}
			}()
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
		return errors.New("swim server is already stopped")
	}

	exit := make(chan error)
	s.stop <- exit

	atomic.SwapUint32(&s.stopped, uint32(1))

	return <-exit
}

// GetMap returns cluster map.
func (s *Server) GetMap() []*swimpb.Member {
	return s.meml.getAll()
}
