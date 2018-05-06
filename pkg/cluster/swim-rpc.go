package cluster

type rpcHandler interface {
	Ping(req *PingMessage, res *Ack) error
	PingRequest(req *PingRequestMessage, res *Ack) error
}

// MethodName indicates what procedure will be called.
type MethodName string

const (
	// Ping : ping
	Ping MethodName = "Ping"
	// PingRequest : request ping
	PingRequest MethodName = "PingRequest"
)

const rpcPrefix string = "Swim"

func (m MethodName) String() string {
	switch m {
	case Ping, PingRequest:
		return rpcPrefix + "." + string(m)
	default:
		return unknown
	}
}

// PingMessage is the ping message of the swim node.
type PingMessage struct {
	CMap CMap
}

// PingRequestMessage is the ping message of the swim node.
type PingRequestMessage struct {
	dstID ID
	CMap  CMap
}

// Ack is the reply message to the ping message.
type Ack struct {
	CMap CMap
}
