package membership

type rpcHandler interface {
	Ping(req *Message, res *Ack) error
	PingRequest(req *Message, res *Ack) error
}

// MethodName indicates what procedure will be called.
type MethodName string

const (
	// Ping : ping
	Ping MethodName = "Ping"
	// PingRequest : request ping
	PingRequest = "PingRequest"
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

// Message is the ping message of the swim node.
type Message struct {
	CMap CMap
}

// Ack is the reply message to the ping message.
type Ack struct {
	CMap CMap
}
