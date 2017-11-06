package swim

import (
	"context"
	"net"
)

// Ping handles received ping message and returns ack.
func (s *Server) Ping(ctx context.Context, in *Ping) (out *Ack, err error) {
	out = &Ack{}

	for _, m := range in.GetMemlist() {
		// set overrides membership list with the given member if the conditions meet.
		s.meml.set(m)
	}

	return out, nil
}

// NextPing returns target node address and ping message.
func (s *Server) NextPing() (addr string, p *Ping) {
	if len(s.q) == 0 {
		s.makeQ()
	}

	for i, m := range s.q {
		if m.Status == Status_FAULTY {
			s.q[i] = nil
			continue
		}

		addr = net.JoinHostPort(m.Addr, m.Port)

		s.q[i] = nil
		s.q = s.q[:i]
		break
	}

	p = &Ping{}
	p.Memlist = append(p.Memlist, s.meml.getAll()...)

	return
}

func (s *Server) makeQ() {
	s.q = s.q[:0]

	s.q = append(s.q, s.meml.getAll()...)
}
