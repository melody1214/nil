package swim

import (
	"net/rpc"
	"sync"
)

// ping sends periodical ping and sends the result through the channel 'pec'.
func (s *Server) ping(pec chan PingError) {
	fetched := s.meml.fetch(1, withNotFaulty(), withNotMyself())
	// Swim cluster has no healthy node.
	if len(fetched) < 1 {
		return
	}

	// Make ping message.
	p := &Message{
		Header:  s.header,
		Members: s.meml.fetch(0),
	}

	// Sends ping message to the target.
	res, err := s.send(Ping, fetched[0].Address, p)
	if err != nil {
		pec <- PingError{
			Type:   Ping,
			DestID: fetched[0].ID,
			Err:    err.Error(),
		}
		return
	}

	for key, value := range s.header {
		f, ok := s.headerFunc[key]
		if ok == false {
			continue
		}

		rcv, ok := res.Header[key]
		if ok == false {
			continue
		}

		if f.compare(value, rcv) == false {
			continue
		}

		// TODO: how if the notiC blocked?
		go func(notiC chan interface{}) {
			notiC <- nil
		}(f.notiC)
	}
}

// pingRequest picks 'k' random member and requests them to send ping 'dstID' indirectly.
func (s *Server) pingRequest(dstID ServerID, pec chan PingError) {
	k := 3                // Number of requests.
	alive := false        // Result of requests.
	var wg sync.WaitGroup // Wait for all requests are finished.

	dst, ok := s.meml.get(dstID)
	if !ok {
		pec <- PingError{
			Type:   PingRequest,
			DestID: dstID,
			Err:    ErrNotFound.Error(),
		}
		return
	}
	content := make([]Member, 1)
	content[0] = dst

	fetched := s.meml.fetch(0, withNotFaulty(), withNotSuspect(), withNotMyself())
	for _, m := range fetched {
		if k == 0 {
			break
		}

		p := &Message{Members: content}

		wg.Add(1)
		go func(addr ServerAddress, ping *Message) {
			defer wg.Done()

			_, err := s.send(PingRequest, addr, ping)
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
		Err:    ErrPingReq.Error(),
	}
}

// send creates gRPC client and send ping by using it.
func (s *Server) send(method MethodName, addr ServerAddress, msg *Message) (ack *Ack, err error) {
	conn, err := s.trans.Dial(string(addr), s.conf.PingExpire)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res := &Ack{}

	cli := rpc.NewClient(conn)
	return res, cli.Call(method.String(), msg, res)
}
