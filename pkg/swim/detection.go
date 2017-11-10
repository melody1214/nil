package swim

import (
	"net"
	"time"

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

func (s *Server) ping() {
	fetched := s.meml.fetch(1)
	// Send ping only the target is not faulty.
	if fetched[0].Status == swimpb.Status_FAULTY {
		return
	}

	// Make ping message.
	p := &swimpb.PingMessage{
		Type:    swimpb.Type_PING,
		Memlist: s.meml.fetch(0),
	}

	// Sends ping message to the target.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.sendPing(ctx, net.JoinHostPort(fetched[0].Addr, fetched[0].Port), p)
	if err != nil {
		// c <- err
		return
	}
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
