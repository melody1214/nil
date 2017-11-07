package swim

import (
	"time"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"golang.org/x/net/context"
)

// Server has functions
type Server struct {
	id   string
	meml *memList

	q []*swimpb.Member
}

// NewServer creates swim server object.
func NewServer(id string, addr, port string) *Server {
	me := newMember(id, addr, port, swimpb.Status_ALIVE, 0)
	memList := newMemList()

	memList.set(me)

	return &Server{
		id:   id,
		meml: memList,
		q:    make([]*swimpb.Member, 0),
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
	for {
		select {
		case <-ticker.C:
			// Get next ping target.
			t, p := s.nextPing()

			// Sends ping message to the target.
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				_, err := s.sendPing(ctx, t, p)
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

func (s *Server) makeQ() {
	s.q = s.q[:0]

	s.q = append(s.q, s.meml.getAll()...)
}
