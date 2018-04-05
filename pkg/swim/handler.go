package swim

import (
	"fmt"
)

// Handler has exposed methods for rpc server.
type Handler struct {
	s *Server
}

func newRPCHandler(s *Server) (RPCHandler, error) {
	if s == nil {
		return nil, fmt.Errorf("nil server object")
	}

	return &Handler{s: s}, nil
}

// Ping is an exposed method of swim rpc service.
// It simply wraps the server's handleJoin method.
func (h *Handler) Ping(req *Message, res *Ack) (err error) {
	return h.s.handlePing(req, res)
}

// PingRequest is an exposed method of swim rpc service.
// It simply wraps the server's handleJoin method.
func (h *Handler) PingRequest(req *Message, res *Ack) (err error) {
	return h.s.handlePingRequest(req, res)
}

// Join is an exposed method of swim rpc service.
// It simply wraps the server's handleJoin method.
func (h *Handler) Join(req *Message, res *Ack) (err error) {
	return h.s.handleJoin(req, res)
}

// handlePing handles received ping message and returns ack.
func (s *Server) handlePing(req *Message, res *Ack) (err error) {
	s.updateMemberList(req.Members)

	return nil
}

// handlePingRequest handling indirect ping request message.
// It pings to the member list in the request message, and return the result.
func (s *Server) handlePingRequest(req *Message, res *Ack) (err error) {
	s.updateMemberList(req.Members)

	for _, m := range req.Members {
		res, err = s.send(Ping, m.Address, req)
		break
	}
	return err
}

// handleJoin handles join request from new node.
func (s *Server) handleJoin(req *Message, res *Ack) (err error) {
	s.updateMemberList(req.Members)

	s.broadcast()
	return nil
}

// updateMemberList checks the update condition of each member in the
// given list. If the conditions are met, then update server's member list.
func (s *Server) updateMemberList(meml []Member) {
	updated := false
	for _, m := range meml {
		if s.meml.set(m) {
			updated = true
		}
	}

	// Sends a notification that the membership list is updated.
	if updated {
		s.c <- PingError{
			Err: ErrChanged.Error(),
		}
	}
}
