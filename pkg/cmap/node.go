package cmap

// Type is the type of the node.
type Type int

const (
	// MDS : metadata server node.
	MDS Type = iota
	// DS : data server node.
	DS
	// GW : gateway node.
	GW
)

// String returns a string of the node type.
func (t Type) String() string {
	switch t {
	case MDS:
		return "MDS"
	case DS:
		return "DS"
	case GW:
		return "GW"
	default:
		return "unknown"
	}
}

// Status is the status of the node.
type Status int

const (
	// Alive : healthy node
	Alive Status = iota
	// Suspect : maybe faulty
	Suspect
	// Faulty : faulty
	Faulty
)

// String returns a string of the node status.
func (s Status) String() string {
	switch s {
	case Alive:
		return "Alive"
	case Suspect:
		return "Suspect"
	case Faulty:
		return "Faulty"
	default:
		return "unknown"
	}
}

// Node is the member of cluster.
type Node struct {
	addr string
	t    Type
	s    Status
}

// Address returns the node address.
func (n *Node) Address() string {
	return n.addr
}

// Type returns the node type.
func (n *Node) Type() Type {
	return n.t
}

// Status returns the node status.
func (n *Node) Status() Status {
	return n.s
}
