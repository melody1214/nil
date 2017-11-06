package mds

import (
	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func (s *server) Ping(ctx context.Context, in *swim.Ping) (out *swim.Ack, err error) {
	return s.swim.Ping(ctx, in)
}

func (s *server) join() error {
	addr, ping := s.swim.NextPing()

	// Test code.
	// Coordinator address
	addr = "127.0.0.1:51000"

	_, err := s.sendPing(context.Background(), addr, ping)
	return err
}

func (s *server) sendPing(ctx context.Context, addr string, ping *swim.Ping) (ack *swim.Ack, err error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := mdspb.NewMdsClient(conn)
	return c.Ping(ctx, ping)
}
