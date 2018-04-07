package cmap

import (
	"fmt"
)

// Version is the version of cluster map.
type Version int64

// Int64 returns built-in int64 type of cmap.Version
func (v Version) Int64() int64 {
	return int64(v)
}

// CMap is a cluster map which includes the information about nodes.
type CMap struct {
	Version Version `xml:"version"`
	Nodes   []Node  `xml:"node"`
}

// HumanReadable returns a human readable map of the cluster.
func (m *CMap) HumanReadable() string {
	out := ""

	// Make human readable sentences for each nodes.
	for _, n := range m.Nodes {
		row := fmt.Sprintf(
			"| %4s | %4s | %s | %7s | %s |\n",
			n.ID.String(),
			n.Type.String(),
			n.Addr,
			n.Stat.String(),
			n.Name,
		)

		out += row
	}

	return out
}

// Save stores the cluster map to the local file system.
func (m *CMap) Save() error {
	return store(m)
}
