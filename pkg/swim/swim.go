package swim

import "context"

// Ping handles received ping message and returns ack.
func (s *Server) Ping(ctx context.Context, in *Ping) (out *Ack, err error) {
	out = &Ack{}

	for _, m := range in.GetMemlist() {
		// set overrides membership list with the given member if the conditions meet.
		s.meml.set(m)
	}

	return nil, nil
}
