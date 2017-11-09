package swim

import (
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"golang.org/x/net/context"
)

// Server has functions
type Server struct {
	id   string
	meml *memList
}

// NewServer creates swim server object.
func NewServer(id string, addr, port string) *Server {
	// Make member myself and add to the list.
	me := newMember(id, addr, port, swimpb.Status_ALIVE, 0)
	memList := newMemList()

	memList.set(me)

	return &Server{
		id:   id,
		meml: memList,
	}
}

// Serve starts gossiping.
func (s *Server) Serve(c chan error) {
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

// GetMap returns cluster map.
func (s *Server) GetMap() []*swimpb.Member {
	return s.meml.getAll()
}
