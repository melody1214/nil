package server

import (
	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"golang.org/x/net/context"
)

// PrintMap returns cluster map info.
func (s *Server) PrintMap(ctx context.Context, in *mdspb.PrintMapRequest) (out *mdspb.PrintMapResponse, err error) {
	return &mdspb.PrintMapResponse{
		Memlist: s.swim.GetMap(),
	}, nil
}
