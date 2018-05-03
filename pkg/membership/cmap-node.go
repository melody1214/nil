package membership

// NodeName is a string for identifying node.
type NodeName string

// NodeAddress is a combination of host:port string of node network address.
type NodeAddress string

// NodeType represents the type of the node.
type NodeType int

const (
	// MDS : Metadata server.
	MDS NodeType = iota
	// DS : Disk server.
	DS
	// GW : Gateway.
	GW
)

// String returns the string of node type.
func (t NodeType) String() string {
	if t == MDS {
		return "MDS"
	} else if t == DS {
		return "DS"
	} else if t == GW {
		return "GW"
	}
	return unknown
}

// NodeStatus is the status of the node.
type NodeStatus string

const (
	// Alive : healthy node
	Alive NodeStatus = "Alive"
	// Suspect : maybe faulty
	Suspect = "Suspect"
	// Faulty : faulty
	Faulty = "Faulty"
)

// String returns a string of the node status.
func (s NodeStatus) String() string {
	switch s {
	case Alive, Suspect, Faulty:
		return string(s)
	default:
		return unknown
	}
}

// Node is the member of cluster.
type Node struct {
	ID   ID          `xml:"id"`
	Name NodeName    `xml:"name"`
	Addr NodeAddress `xml:"address"`
	Type NodeType    `xml:"type"`
	Stat NodeStatus  `xml:"status"`
	Vols []ID        `xml:"volume"`
}
