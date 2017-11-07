package swim

import (
	"net"

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

	return out, nil
}

// nextPing returns target node address and ping message.
func (s *Server) nextPing() (addr string, p *swimpb.PingMessage) {
	if len(s.q) == 0 {
		s.makeQ()
	}

	for i, m := range s.q {
		if m.Status == swimpb.Status_FAULTY {
			s.q[i] = nil
			continue
		}

		addr = net.JoinHostPort(m.Addr, m.Port)

		s.q[i] = nil
		s.q = s.q[:i]
		break
	}

	p = &swimpb.PingMessage{}
	p.Memlist = append(p.Memlist, s.meml.getAll()...)

	return
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
