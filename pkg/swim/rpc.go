package swim

// RPCHandler is the interface of swim rpc commands.
type RPCHandler interface {
	Ping(req *Message, res *Ack) error
	PingRequest(req *Message, res *Ack) error
	Broadcast(req *Message, res *Ack) error
}

// MethodName indicates what procedure will be called.
type MethodName int

const (
	// Ping : ping
	Ping MethodName = iota
	// PingRequest : request ping
	PingRequest
	// Broadcast : request boradcasting
	Broadcast
)

const rpcPrefix string = "Server"

func (m MethodName) String() string {
	switch m {
	case Ping:
		return rpcPrefix + "." + "Ping"
	case PingRequest:
		return rpcPrefix + "." + "PingRequest"
	case Broadcast:
		return rpcPrefix + "." + "Broadcast"
	default:
		return "unknown"
	}
}

// Message is the basic message of the swim node.
type Message struct {
	Members []Member
}

// Ack is the reply message to the SwimMessage.
type Ack struct{}
