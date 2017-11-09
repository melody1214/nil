package swim

import (
	"context"
	"net"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

// Join try to join the membership.
func (s *Server) join() error {
	// Test code.
	// Coordinator address
	addr := "127.0.0.1:51000"

	// Make ping message.
	ping := &swimpb.PingMessage{}
	ping.Type = swimpb.Type_BROADCAST
	ping.Memlist = append(ping.Memlist, s.meml.getAll()...)

	// Send join message to coordinator.
	_, err := s.sendPing(context.Background(), addr, ping)
	return err
}

// Send ping message to all.
func (s *Server) broadcast() {
	meml := s.meml.getAll()

	for _, m := range meml {
		if m.Status == swimpb.Status_FAULTY {
			continue
		}

		p := &swimpb.PingMessage{
			Type: swimpb.Type_PING,
		}
		p.Memlist = append(p.Memlist, meml...)

		go func() {
			s.sendPing(context.Background(),
				net.JoinHostPort(m.Addr, m.Port),
				p)
		}()
	}
}
