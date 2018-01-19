package swim

// Ping handles received ping message and returns ack.
func (s *Server) Ping(req *Message, res *Ack) (err error) {
	s.updateMemberList(req.Members)

	return nil
}

// PingRequest handling indirect ping request message.
// It pings to the member list in the request message, and return the result.
func (s *Server) PingRequest(req *Message, res *Ack) (err error) {
	s.updateMemberList(req.Members)

	for _, m := range req.Members {
		res, err = s.send(Ping, m.Address, req)
		break
	}
	return err
}

// Join handles join request from new node.
func (s *Server) Join(req *Message, res *Ack) (err error) {
	s.updateMemberList(req.Members)

	s.broadcast()
	return nil
}

// updateMemberList checks the update condition of each member in the
// given list. If the conditions are met, then update server's member list.
func (s *Server) updateMemberList(meml []Member) {
	for _, m := range meml {
		s.meml.set(m)
	}
}
