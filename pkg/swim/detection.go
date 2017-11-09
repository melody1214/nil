package swim

import (
	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Ping handles received ping message and returns ack.
func (s *Server) Ping(ctx context.Context, in *swimpb.PingMessage) (out *swimpb.Ack, err error) {
	out = &swimpb.Ack{}

	for _, m := range in.GetMemlist() {
		// set overrides membership list with the given member if the conditions meet.
		s.meml.set(m)
	}

	switch in.GetType() {
	case swimpb.Type_BROADCAST:
		s.broadcast()
	}

	return out, nil
}

func (s *Server) sendPing(ctx context.Context, addr string, ping *swimpb.PingMessage) (ack *swimpb.Ack, err error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := swimpb.NewSwimClient(conn)
	return c.Ping(ctx, ping)
}
