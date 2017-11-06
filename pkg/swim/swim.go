package swim

import (
	"context"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

// Ping handles received ping message and returns ack.
func (s *Server) Ping(ctx context.Context, in *Ping) (out *Ack, err error) {
	out = &Ack{}

	for _, m := range in.GetMemlist() {
		// set overrides membership list with the given member if the conditions meet.
		s.meml.set(newMember(
			uuid.UUID(m.Uuid),
			m.Addr,
			m.Port,
			int32(m.Status),
			m.Incarnation,
		))
	}

	return nil, nil
}
