package swim

import (
	"context"
	"net"
	"runtime"
	"time"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

// Join try to join the membership.
func (s *Server) join() error {
	// Test code. Coordinator address.
	coordinator := "127.0.0.1:51000"

	return s.askBroadcast(coordinator, s.meml.fetch(0))
}

// Leave try to send leaving message to all members.
func (s *Server) leave() error {
	return s.disseminate(s.id, swimpb.Status_FAULTY)
}

// Alive try to send alive message to all members.
func (s *Server) alive(id string) error {
	return s.disseminate(id, swimpb.Status_ALIVE)
}

// Suspect try to send suspect message to all members.
func (s *Server) suspect(id string) error {
	return s.disseminate(id, swimpb.Status_SUSPECT)
}

// Faulty try to send faulty message to all members.
func (s *Server) faulty(id string) error {
	return s.disseminate(id, swimpb.Status_FAULTY)
}

// Disseminate change the status and ask broadcast it to other healthy node.
func (s *Server) disseminate(id string, status swimpb.Status) error {
	// Change status.
	m := s.meml.get(id)
	m.Status = status

	// If changing my status, then increase incarnation number.
	if s.id == id {
		m.Incarnation++
	}

	// Update information.
	s.meml.set(m)

	// Prepare ping message content.
	content := make([]*swimpb.Member, 1)
	content[0] = m

	// Choose gossiper.
	meml := s.meml.fetch(0)
	var gossiper *swimpb.Member
	for _, m := range meml {
		// Gossiper would be healthy and not myself.
		if m.Uuid != s.id && m.Status == swimpb.Status_ALIVE {
			gossiper = m
			break
		}
	}

	// I'm the only survivor of this membership.
	// There is no member who is able to gossip my message.
	if gossiper == nil {
		return nil
	}

	return s.askBroadcast(net.JoinHostPort(gossiper.Addr, gossiper.Port), content)
}

// askBroadcast ask gossiper to broadcast message.
func (s *Server) askBroadcast(gossiper string, meml []*swimpb.Member) error {
	ping := &swimpb.PingMessage{
		Type:    swimpb.Type_BROADCAST,
		Memlist: meml,
	}

	_, err := s.sendPing(context.Background(), gossiper, ping)
	return err
}

// Send ping message to all.
func (s *Server) broadcast() {
	ml := s.meml.fetch(0)

	for _, m := range ml {
		if m.Status == swimpb.Status_FAULTY {
			continue
		}

		p := &swimpb.PingMessage{
			Type:    swimpb.Type_PING,
			Memlist: ml,
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			s.sendPing(ctx, net.JoinHostPort(m.Addr, m.Port), p)
		}()
	}
	runtime.Gosched()
}
