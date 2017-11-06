package mds

import (
	"github.com/chanyoung/nil/pkg/swim"
	"golang.org/x/net/context"
)

func (s *server) Ping(ctx context.Context, in *swim.Ping) (out *swim.Ack, err error) {
	return s.swim.Ping(ctx, in)
}
