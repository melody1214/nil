package swim

import (
	"context"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

// Join try to join the membership.
func (s *Server) join() error {
	// Test code. // Coordinator address
	addr := "127.0.0.1:51000"

	// Make ping message.
	ping := &swimpb.PingMessage{}
	ping.Type = swimpb.Type_BROADCAST
	ping.Memlist = append(ping.Memlist, s.meml.getAll()...)

	// Send join message to coordinator.
	_, err := s.sendPing(context.Background(), addr, ping)
	return err
}

// Leave try to leave the membership by it's will.
func (s *Server) leave() {
	s.meml.changeStatus(s.id, swimpb.Status_FAULTY)
	s.broadcast()
}

// Send ping message to all.
func (s *Server) broadcast() {
	meml := s.meml.getAll()

	var wg sync.WaitGroup
	for _, m := range meml {
		if m.Status == swimpb.Status_FAULTY {
			continue
		}

		p := &swimpb.PingMessage{
			Type: swimpb.Type_PING,
		}
		p.Memlist = append(p.Memlist, meml...)

		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			s.sendPing(ctx, net.JoinHostPort(m.Addr, m.Port), p)
		}()
		runtime.Gosched()
	}
	wg.Wait()
}
