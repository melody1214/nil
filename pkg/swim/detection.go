package swim

import (
	"net"
	"sync"

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
	case swimpb.Type_PING:
		return out, nil

	case swimpb.Type_BROADCAST:
		s.broadcast()
		return out, nil

	case swimpb.Type_PINGREQUEST:
		meml := in.GetMemlist()
		if len(meml) == 0 {
			return out, ErrNotFound
		}
		return s.sendPing(ctx, net.JoinHostPort(meml[0].Addr, meml[0].Port), in)

	default:
		return out, nil
	}
}

// ping sends periodical ping and sends the result through the channel 'pec'.
func (s *Server) ping(pec chan PingError) {
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
	ctx, cancel := context.WithTimeout(context.Background(), pingExpire)
	defer cancel()

	_, err := s.sendPing(ctx, net.JoinHostPort(fetched[0].Addr, fetched[0].Port), p)
	if err != nil {
		pec <- PingError{
			Type:   swimpb.Type_PING,
			DestID: fetched[0].Uuid,
			Err:    err,
		}
		return
	}
}

// pingRequest picks 'k' random member and requests them to send ping 'dstID' indirectly.
func (s *Server) pingRequest(dstID string, pec chan PingError) {
	k := 3                // Number of requests.
	alive := false        // Result of requests.
	var wg sync.WaitGroup // Wait for all requests are finished.

	dst := s.meml.get(dstID)
	if dst == nil {
		pec <- PingError{
			Type:   swimpb.Type_PINGREQUEST,
			DestID: dstID,
			Err:    ErrNotFound,
		}
		return
	}
	content := make([]*swimpb.Member, 1)
	content[0] = dst

	fetched := s.meml.fetch(0)
	for _, m := range fetched {
		if k == 0 {
			break
		}

		if m.Status != swimpb.Status_ALIVE {
			continue
		}

		if m.Uuid == s.id {
			continue
		}

		p := &swimpb.PingMessage{
			Type:    swimpb.Type_PINGREQUEST,
			Memlist: content,
		}

		wg.Add(1)
		go func(addr string, ping *swimpb.PingMessage) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), pingExpire)
			defer cancel()

			_, err := s.sendPing(ctx, addr, ping)
			if err == nil {
				alive = true
			}
		}(net.JoinHostPort(m.Addr, m.Port), p)

		k--
	}

	wg.Wait()

	if alive {
		s.alive(dstID)
		return
	}

	pec <- PingError{
		Type:   swimpb.Type_PINGREQUEST,
		DestID: dstID,
		Err:    ErrPingReq,
	}
}

// sendPing creates gRPC client and send ping by using it.
func (s *Server) sendPing(ctx context.Context, addr string, ping *swimpb.PingMessage) (ack *swimpb.Ack, err error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := swimpb.NewSwimClient(conn)
	return c.Ping(ctx, ping)
}
