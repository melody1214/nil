package membership

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Incarnation is the versioning information and using instead of time.
type Incarnation uint32

// Uint32 returns its value in built-in type uint32.
func (i Incarnation) Uint32() uint32 {
	return uint32(i)
}

// ping sends periodical ping and sends the result through the channel 'pec'.
func (s *server) ping() {
	cmap := s.cMapManager.LatestCMap()

	const notFound = NodeName("not found")
	fetched := Node{Name: notFound}

	randIdx := s.cMapManager.random.Perm(len(cmap.Nodes))
	for i := 0; i < len(cmap.Nodes); i++ {
		n := cmap.Nodes[randIdx[i]]

		// Not myself.
		if n.Name == s.cfg.Name {
			continue
		}

		// Not faulty node.
		if n.Stat == Faulty {
			continue
		}

		// Not gateway.
		if n.Type == GW {
			continue
		}

		fetched = n
	}

	if fetched.Name == notFound {
		// No ping available node.
		return
	}

	msg := &Message{CMap: cmap}

	ack, err := s.send(Ping, fetched.Addr, msg)
	if err != nil {
		logger.Warn(errors.Wrapf(err, "failed to send ping message to node %+v", fetched))

		// Do disseminate
		// Wait for a minute and retry via ping request.
		time.Sleep(1 * time.Minute)
		s.pingRequest(fetched.ID)
		return
	}

	s.cMapManager.mergeCMap(&ack.CMap)
}

// pingRequest picks 'k' random member and requests them to send ping 'dstID' indirectly.
func (s *server) pingRequest(dstID ID) {
	k := 3                // Number of requests.
	alive := false        // Result of requests.
	var wg sync.WaitGroup // Wait for all requests are finished.

	dstNode, err := s.cMapManager.SearchCallNode().ID(dstID).Do()
	if err != nil {
		logger.Error(errors.Wrapf(err, "failed to find ping request destination node: %v", dstID))
		// Do faulty
		return
	}
	if dstNode.Stat != Suspect {
		// Already handled.
		return
	}

	cmap := s.cMapManager.LatestCMap()
	randIdx := s.cMapManager.random.Perm(len(cmap.Nodes))
	for i := 0; i < len(cmap.Nodes); i++ {
		n := cmap.Nodes[randIdx[i]]

		// Not myself.
		if n.Name == s.cfg.Name {
			continue
		}

		// Not faulty or suspect node.
		if n.Stat == Faulty || n.Stat == Suspect {
			continue
		}

		// Not gateway.
		if n.Type == GW {
			continue
		}

		wg.Add(1)
		go func(addr NodeAddress, msg *Message) {
			defer wg.Done()

			_, err := s.send(PingRequest, addr, msg)
			if err == nil {
				alive = true
			}
		}(n.Addr, &Message{CMap: cmap})
		k--

		if k == 0 {
			break
		}
	}

	wg.Wait()

	if alive {
		// Do alive
		return
	}

	s.cMapManager.Outdated()
	// Do disseminate
}

// send creates rpc client and send ping message by using it.
func (s *server) send(method MethodName, addr NodeAddress, msg *Message) (ack *Ack, err error) {
	conn, err := s.trans.Dial(addr.String(), s.cfg.PingExpire)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dialing")
	}
	defer conn.Close()

	res := &Ack{}
	cli := rpc.NewClient(conn)
	return res, cli.Call(method.String(), msg, res)
}

// Ping handles ping request.
func (s *server) Ping(req *Message, res *Ack) (err error) {
	return nil
}

// PingRequest handles ping-request request.
func (s *server) PingRequest(req *Message, res *Ack) (err error) {
	return nil
}
