package server

import (
	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/security"
	"golang.org/x/net/context"
)

// PrintMap returns cluster map info.
func (s *Server) GetClusterMap(ctx context.Context, in *mdspb.GetClusterMapRequest) (out *mdspb.GetClusterMapResponse, err error) {
	return &mdspb.GetClusterMapResponse{
		Memlist: s.swim.GetMap(),
	}, nil
}

// AddUser creates user and returns API key.
func (s *Server) AddUser(ctx context.Context, in *mdspb.AddUserRequest) (out *mdspb.AddUserResponse, err error) {
	ak := security.NewAPIKey()

	return &mdspb.AddUserResponse{
		AccessKey: ak.AccessKey(),
		SecretKey: ak.SecretKey(),
	}, nil
}
