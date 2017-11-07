package swim

import "context"

// Join try to join the membership.
func (s *Server) join() error {
	addr, ping := s.nextPing()

	// Test code.
	// Coordinator address
	addr = "127.0.0.1:51000"

	_, err := s.sendPing(context.Background(), addr, ping)
	return err
}
