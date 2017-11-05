package mds

import (
	"github.com/chanyoung/nil/pkg/swim"
	"golang.org/x/net/context"
)

func (s *server) Ping(ctx context.Context, in *swim.Ping) (out *swim.Ack, err error) {
	out = &swim.Ack{}

	for _, m := range in.GetMemlist() {
		s.swim.Set(m)
	}

	return nil, nil
}
