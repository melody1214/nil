package cmap

const unknown = "Unknown"

// EncodingGroupStatus is the status of the node.
type EncodingGroupStatus string

const (
	// EGAlive : healthy node
	EGAlive EncodingGroupStatus = "Alive"
	// EGSuspect : maybe faulty
	EGSuspect EncodingGroupStatus = "Suspect"
	// EGFaulty : faulty
	EGFaulty EncodingGroupStatus = "Faulty"
	// EGRdonly : readonly, maybe rebalancing or recovering.
	EGRdonly EncodingGroupStatus = "Rdonly"
)

// String returns a string of the node status.
func (s EncodingGroupStatus) String() string {
	switch s {
	case EGAlive, EGSuspect, EGFaulty, EGRdonly:
		return string(s)
	default:
		return unknown
	}
}

// EncodingGroup is the logical group for making local parity.
type EncodingGroup struct {
	ID   ID                  `xml:"id"`
	Incr Incarnation         `xml:"incarnation"`
	Stat EncodingGroupStatus `xml:"status"`
	Size int64               `xml:"size"`
	Used int64               `xml:"used"`
	Free int64               `xml:"free"`
	Vols []ID                `xml:"volume"`
	Uenc int                 `xml:"unencoded"`
}

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
	// Alive : healthy node
	Alive NodeStatus = "Alive"
	// Suspect : maybe faulty
	Suspect NodeStatus = "Suspect"
	// Faulty : faulty
	Faulty NodeStatus = "Faulty"
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
	Incr Incarnation `xml:"incarnation"`
	Name NodeName    `xml:"name"`
	Addr NodeAddress `xml:"address"`
	Type NodeType    `xml:"type"`
	Stat NodeStatus  `xml:"status"`
	Vols []ID        `xml:"volume"`
}

// VolumeSpeed represents a disk speed level.
type VolumeSpeed string

const (
	// Low : 60 Mb/s < speed < 80 Mb/s
	Low VolumeSpeed = "Low"
	// Mid : 80 Mb/s < speed < 100 Mb/s
	Mid VolumeSpeed = "Mid"
	// High : 100 Mb/s < speed
	High VolumeSpeed = "High"
)

func (s VolumeSpeed) String() string {
	switch s {
	case Low, Mid, High:
		return string(s)
	default:
		return unknown
	}
}

// VolumeStatus represents a disk status.
type VolumeStatus string

const (
	// Prepared represents the volume is ready to run.
	Prepared VolumeStatus = "Prepared"
	// Active represents the volume is now running.
	Active VolumeStatus = "Active"
	// Failed represents the volume has some problems and stopped now.
	Failed VolumeStatus = "Failed"
)

func (s VolumeStatus) String() string {
	switch s {
	case Prepared, Active, Failed:
		return string(s)
	default:
		return unknown
	}
}

// Volume is volumes which is attached in the ds.
type Volume struct {
	ID      ID           `xml:"id"`
	Incr    Incarnation  `xml:"incarnation"`
	Size    uint64       `xml:"size"`
	Speed   VolumeSpeed  `xml:"speed"`
	Stat    VolumeStatus `xml:"status"`
	Node    ID           `xml:"node"`
	EncGrps []ID         `xml:"encgrp"`
	MaxEG   int          `xml:"maxeg"`
}
