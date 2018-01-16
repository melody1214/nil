package swim

import (
	"net/rpc"
	"sync"
)

// Ping handles received ping message and returns ack.
func (s *Server) Ping(req *Message, res *Ack) (err error) {
	for _, m := range req.Members {
		// set overrides membership list with the given member if the conditions meet.
		s.meml.set(m)
	}

	switch req.Type {
	case Ping:
		return nil

	case Broadcast:
		s.broadcast()
		return nil

	case PingRequest:
		meml := req.Members
		if len(meml) == 0 {
			return ErrNotFound
		}
		res, err = s.sendPing(meml[0].Address, req)
		return err

	default:
		return nil
	}
}

// ping sends periodical ping and sends the result through the channel 'pec'.
func (s *Server) ping(pec chan PingError) {
	fetched := s.meml.fetch(1)
	// Send ping only the target is not faulty.
	if fetched[0].Status == Faulty {
		return
	}

	// Make ping message.
	p := &Message{
		Type:    Ping,
		Members: s.meml.fetch(0),
	}

	// Sends ping message to the target.
	_, err := s.sendPing(fetched[0].Address, p)
	if err != nil {
		pec <- PingError{
			Type:   Ping,
			DestID: fetched[0].ID,
			Err:    err,
		}
		return
	}
}

// pingRequest picks 'k' random member and requests them to send ping 'dstID' indirectly.
func (s *Server) pingRequest(dstID ServerID, pec chan PingError) {
	k := 3                // Number of requests.
	alive := false        // Result of requests.
	var wg sync.WaitGroup // Wait for all requests are finished.

	dst := s.meml.get(dstID)
	if dst == nil {
		pec <- PingError{
			Type:   PingRequest,
			DestID: dstID,
			Err:    ErrNotFound,
		}
		return
	}
	content := make([]*Member, 1)
	content[0] = dst

	fetched := s.meml.fetch(0)
	for _, m := range fetched {
		if k == 0 {
			break
		}

		if m.Status != Alive {
			continue
		}

		if m.ID == s.conf.ID {
			continue
		}

		p := &Message{
			Type:    PingRequest,
			Members: content,
		}

		wg.Add(1)
		go func(addr ServerAddress, ping *Message) {
			defer wg.Done()

			_, err := s.sendPing(addr, ping)
			if err == nil {
				alive = true
			}
		}(m.Address, p)

		k--
	}

	wg.Wait()

	if alive {
		s.alive(dstID)
		return
	}

	pec <- PingError{
		Type:   PingRequest,
		DestID: dstID,
		Err:    ErrPingReq,
	}
}

// sendPing creates gRPC client and send ping by using it.
func (s *Server) sendPing(addr ServerAddress, ping *Message) (ack *Ack, err error) {
	conn, err := s.trans.Dial(string(addr), s.conf.PingExpire)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res := &Ack{}

	cli := rpc.NewClient(conn)
	return res, cli.Call("Server.Ping", ping, res)
}
