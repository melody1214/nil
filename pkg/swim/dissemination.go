package swim

import (
	"context"
	"net"
	"runtime"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

// Join tries to join the membership.
func (s *Server) join() error {
	// Test code. Coordinator address.
	coordinator := s.cfg.ClusterJoinAddr

	return s.askBroadcast(coordinator, s.meml.fetch(0))
}

// Leave tries to send leaving message to all members.
func (s *Server) leave() error {
	return s.disseminate(s.cfg.ID, swimpb.Status_FAULTY)
}

// Alive tries to send alive message to all members.
func (s *Server) alive(id string) error {
	return s.disseminate(id, swimpb.Status_ALIVE)
}

// Suspect tries to send suspect message to all members.
func (s *Server) suspect(id string) error {
	return s.disseminate(id, swimpb.Status_SUSPECT)
}

// Faulty tries to send faulty message to all members.
func (s *Server) faulty(id string) error {
	return s.disseminate(id, swimpb.Status_FAULTY)
}

// Disseminate changes the status and asks broadcast it to other healthy node.
func (s *Server) disseminate(id string, status swimpb.Status) error {
	// Change status.
	m := s.meml.get(id)
	m.Status = status

	// If changing my status, then increase incarnation number.
	if s.cfg.ID == id {
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
		if m.Uuid != s.cfg.ID && m.Status == swimpb.Status_ALIVE {
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

// askBroadcast asks broadcasting message to gossiper node.
func (s *Server) askBroadcast(gossiper string, meml []*swimpb.Member) error {
	ping := &swimpb.PingMessage{
		Type:    swimpb.MessageType_BROADCAST,
		Memlist: meml,
	}

	_, err := s.sendPing(context.Background(), gossiper, ping)
	return err
}

// broadcast sends ping message to all.
func (s *Server) broadcast() {
	ml := s.meml.fetch(0)

	for _, m := range ml {
		if m.Status == swimpb.Status_FAULTY {
			continue
		}

		p := &swimpb.PingMessage{
			Type:    swimpb.MessageType_PING,
			Memlist: ml,
		}

		go func(addr string, ping *swimpb.PingMessage) {
			ctx, cancel := context.WithTimeout(context.Background(), pingExpire)
			defer cancel()

			s.sendPing(ctx, addr, ping)
		}(net.JoinHostPort(m.Addr, m.Port), p)
	}
	runtime.Gosched()
}
