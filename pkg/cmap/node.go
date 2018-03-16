package cmap

import (
	"strconv"
)

// unknown: unknown type or status of node.
const unknown string = "unknown"

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
		return unknown
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
		return unknown
	}
}

// ID is the id of the node.
type ID int64

// String returns a string type of the ID.
func (i ID) String() string {
	return strconv.FormatInt(i.Int64(), 10)
}

// Int64 returns a int64 type of the ID.
func (i ID) Int64() int64 {
	return int64(i)
}

// Node is the member of cluster.
type Node struct {
	ID   ID     `xml:"id"`
	Name string `xml:"name"`
	Addr string `xml:"address"`
	Type Type   `xml:"type"`
	Stat Status `xml:"status"`
}
