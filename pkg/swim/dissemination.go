package swim

import (
	"runtime"
)

// Join tries to join the membership.
func (s *Server) join() error {
	// Test code. Coordinator address.
	return s.askBroadcast(s.conf.Coordinator, s.meml.fetch(0))
}

// Leave tries to send leaving message to all members.
func (s *Server) leave() error {
	return s.disseminate(s.conf.ID, Faulty)
}

// Alive tries to send alive message to all members.
func (s *Server) alive(id ServerID) error {
	return s.disseminate(id, Alive)
}

// Suspect tries to send suspect message to all members.
func (s *Server) suspect(id ServerID) error {
	return s.disseminate(id, Suspect)
}

// Faulty tries to send faulty message to all members.
func (s *Server) faulty(id ServerID) error {
	return s.disseminate(id, Faulty)
}

// Disseminate changes the status and asks broadcast it to other healthy node.
func (s *Server) disseminate(id ServerID, status Status) error {
	// Change status.
	m, ok := s.meml.get(id)
	if !ok {
		return ErrNotFound
	}

	m.Status = status
	// If changing my status, then increase incarnation number.
	if s.conf.ID == id {
		m.Incarnation++
	}

	// Update information.
	s.meml.set(m)

	// Prepare ping message content.
	content := make([]Member, 1)
	content[0] = m

	// Choose gossiper.
	gossiper := s.meml.fetch(1, withNotFaulty(), withNotSuspect(), withNotMyself())

	// I'm the only survivor of this membership.
	// There is no member who is able to gossip my message.
	if len(gossiper) < 1 {
		return nil
	}

	return s.askBroadcast(gossiper[0].Address, content)
}

// askBroadcast asks broadcasting message to gossiper node.
func (s *Server) askBroadcast(gossiper ServerAddress, meml []Member) error {
	ping := &Message{
		Members: meml,
	}

	_, err := s.send(Broadcast, gossiper, ping)
	return err
}

// broadcast sends ping message to all.
func (s *Server) broadcast() {
	ml := s.meml.fetch(0, withNotFaulty(), withNotSuspect())

	for _, m := range ml {
		p := &Message{Members: ml}
		go s.send(Ping, m.Address, p)
	}
	runtime.Gosched()
}
