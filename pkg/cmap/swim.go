package cmap

import (
	"net/rpc"
	"runtime"
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
	cmap := s.manager.LatestCMap()

	const notFound = NodeName("not found")
	fetched := Node{Name: notFound}

	randIdx := random.Perm(len(cmap.Nodes))
	for i := 0; i < len(cmap.Nodes); i++ {
		n := cmap.Nodes[randIdx[i]]

		// Not myself.
		if n.Name == s.cfg.Name {
			continue
		}

		// Not faulty node.
		if n.Stat == NodeFaulty {
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

	msg := &PingMessage{CMap: *cmap}

	ack, err := s.sendPing(fetched.Addr, msg)
	if err != nil {
		logger.Warn(errors.Wrapf(err, "failed to send ping message to node %+v", fetched))

		s.disseminate(fetched.ID, NodeSuspect)
		// Wait for a minute and retry via ping request.
		time.Sleep(1 * time.Minute)
		s.pingRequest(fetched.ID)
		return
	}

	s.manager.mergeCMap(&ack.CMap)
}

// pingRequest picks 'k' random member and requests them to send ping 'dstID' indirectly.
func (s *server) pingRequest(dstID ID) {
	k := 3                // Number of requests.
	alive := false        // Result of requests.
	var wg sync.WaitGroup // Wait for all requests are finished.

	dstNode, err := s.manager.SearchCall().Node().ID(dstID).Do()
	if err != nil {
		logger.Error(errors.Wrapf(err, "failed to find ping request destination node: %v", dstID))
		s.disseminate(dstID, NodeFaulty)
		return
	}
	if dstNode.Stat != NodeSuspect {
		// Already handled.
		return
	}

	cmap := s.manager.LatestCMap()
	randIdx := random.Perm(len(cmap.Nodes))
	for i := 0; i < len(cmap.Nodes); i++ {
		n := cmap.Nodes[randIdx[i]]

		// Not myself.
		if n.Name == s.cfg.Name {
			continue
		}

		// Not faulty or suspect node.
		if n.Stat == NodeFaulty || n.Stat == NodeSuspect {
			continue
		}

		// Not gateway.
		if n.Type == GW {
			continue
		}

		wg.Add(1)
		go func(addr NodeAddress, msg *PingRequestMessage) {
			defer wg.Done()

			_, err := s.sendPingRequest(addr, msg)
			if err == nil {
				alive = true
			}
		}(n.Addr, &PingRequestMessage{dstID: dstID, CMap: *cmap})
		k--

		if k == 0 {
			break
		}
	}

	wg.Wait()

	if alive {
		// NodeSuspected node will make themselves to alive.
		return
	}

	s.disseminate(dstID, NodeFaulty)
}

// sendPing creates rpc client and send ping message by using it.
func (s *server) sendPing(addr NodeAddress, msg *PingMessage) (ack *Ack, err error) {
	conn, err := s.trans.Dial(addr.String(), s.cfg.PingExpire)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dialing")
	}
	defer conn.Close()

	res := &Ack{}
	cli := rpc.NewClient(conn)
	return res, cli.Call(Ping.String(), msg, res)
}

// sendPingRequest creates rpc client and send ping message by using it.
func (s *server) sendPingRequest(addr NodeAddress, msg *PingRequestMessage) (ack *Ack, err error) {
	conn, err := s.trans.Dial(addr.String(), s.cfg.PingExpire)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dialing")
	}
	defer conn.Close()

	res := &Ack{}
	cli := rpc.NewClient(conn)
	return res, cli.Call(PingRequest.String(), msg, res)
}

// Ping handles ping request.
func (s *server) Ping(req *PingMessage, res *Ack) (err error) {
	s.manager.mergeCMap(&req.CMap)
	res.CMap = *(s.manager.LatestCMap())
	return nil
}

// PingRequest handles ping-request request.
func (s *server) PingRequest(req *PingRequestMessage, res *Ack) (err error) {
	s.manager.mergeCMap(&req.CMap)
	n, err := s.manager.SearchCall().Node().ID(req.dstID).Do()
	if err != nil {
		return err
	}
	ack, err := s.sendPing(n.Addr, &PingMessage{CMap: *(s.manager.LatestCMap())})
	if err != nil {
		return err
	}
	s.manager.mergeCMap(&ack.CMap)
	res.CMap = ack.CMap
	return nil
}

// leave set myself faulty and send it to the cluster.
func (s *server) leave() {
	n, err := s.manager.SearchCall().Node().Name(s.cfg.Name).Do()
	if err != nil {
		return
	}

	s.disseminate(n.ID, NodeFaulty)
}

// Disseminate changes the status and asks broadcast it to other healthy node.
func (s *server) disseminate(id ID, stat NodeStatus) {
	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	cmap := s.manager.latestCMap()
	for i, n := range cmap.Nodes {
		if n.ID != id {
			continue
		}

		cmap.Nodes[i].Stat = stat
		if n.Name == s.cfg.Name {
			cmap.Nodes[i].Incr++
		}

		s.manager.sendStateChangedNotiToAll()
		go s.broadcast()
		return
	}
}

// broadcast sends ping message to all.
func (s *server) broadcast() {
	// Randomly select 3 nodes and send.
	// Too hard to broadcast without IP multicast.
	cmap := s.manager.LatestCMap()
	for k := 0; k < 3; k++ {
		n, err := s.manager.SearchCall().Node().Status(NodeAlive).Random().Do()
		if err != nil {
			continue
		}
		if n.Type == GW || n.Name == s.cfg.Name {
			continue
		}
		go s.sendPing(n.Addr, &PingMessage{CMap: *cmap})
	}
	runtime.Gosched()
}
