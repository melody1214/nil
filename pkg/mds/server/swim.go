package server

// JoinMembership try to join the membership via swim protocol.
func (s *Server) JoinMembership() error {
	return s.swim.Join()
}
