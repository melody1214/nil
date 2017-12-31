package swim

// RPCHandler is the interface of swim rpc commands.
type RPCHandler interface {
	Ping(req *Message, res *Ack) error
}

// MessageType indicates the type of the swim protocol messages.
type MessageType int

const (
	// Ping : ping message
	Ping MessageType = 0
	// PingRequest : request ping message
	PingRequest = 1
	// Broadcast : request boradcasting message
	Broadcast = 2
)

// Message is the basic message of the swim node.
type Message struct {
	Type    MessageType
	Members []*Member
}

// Ack is the reply message to the SwimMessage.
type Ack struct{}
