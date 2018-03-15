package cmap

// Type is the type of the node.
type Type string

const (
	// MDS : metadata server node.
	MDS Type = "MDS"
	// DS : data server node.
	DS = "DS"
	// GW : gateway node.
	GW = "GW"
)

// String returns a string of the node type.
func (t Type) String() string {
	switch t {
	case MDS, DS, GW:
		return string(t)
	default:
		return "unknown"
	}
}

// Status is the status of the node.
type Status string

const (
	// Alive : healthy node
	Alive Status = "Alive"
	// Suspect : maybe faulty
	Suspect = "Suspect"
	// Faulty : faulty
	Faulty = "Faulty"
)

// String returns a string of the node status.
func (s Status) String() string {
	switch s {
	case Alive, Suspect, Faulty:
		return string(s)
	default:
		return "unknown"
	}
}

// Node is the member of cluster.
type Node struct {
	Addr string `xml:"address"`
	Type Type   `xml:"type"`
	Stat Status `xml:"status"`
}
