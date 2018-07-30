package cmap

const unknown = "Unknown"

// NodeName is a string for identifying node.
type NodeName string

func (n NodeName) String() string { return string(n) }

// NodeAddress is a combination of host:port string of node network address.
type NodeAddress string

// String returns its value converted to string type.
func (a NodeAddress) String() string { return string(a) }

// NodeType represents the type of the node.
type NodeType string

const (
	// MDS : Metadata server.
	MDS NodeType = "MDS"
	// DS : Disk server.
	DS NodeType = "DS"
	// GW : Gateway.
	GW NodeType = "GW"
)

// String returns the string of node type.
func (t NodeType) String() string {
	switch t {
	case MDS, DS, GW:
		return string(t)
	default:
		return unknown
	}
}

// NodeStatus is the status of the node.
type NodeStatus string

const (
	// NodeAlive : healthy node
	NodeAlive NodeStatus = "Alive"
	// NodeSuspect : maybe faulty
	NodeSuspect NodeStatus = "Suspect"
	// NodeFaulty : faulty
	NodeFaulty NodeStatus = "Faulty"
)

// String returns a string of the node status.
func (s NodeStatus) String() string {
	switch s {
	case NodeAlive, NodeSuspect, NodeFaulty:
		return string(s)
	default:
		return unknown
	}
}

// Node is the member of cluster.
type Node struct {
	ID   ID          `xml:"id"`
	Incr Incarnation `xml:"incarnation"`
	Name NodeName    `xml:"name"`
	Addr NodeAddress `xml:"address"`
	Type NodeType    `xml:"type"`
	Stat NodeStatus  `xml:"status"`
}
