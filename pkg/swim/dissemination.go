package swim

import (
	"runtime"
)

// Join tries to join the membership.
func (s *Server) join() error {
	_, err := s.send(Join, s.conf.Coordinator, &Message{s.meml.fetch(0)})
	return err
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

	// Send to all.
	s.broadcast()

	return nil
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
