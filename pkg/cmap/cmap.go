package cmap

import (
	"fmt"
)

// CMap is a cluster map which includes the information about nodes.
type CMap struct {
	Version int64  `xml:"version"`
	Nodes   []Node `xml:"node"`
}

// NewCMap returns an empty cluster map.
func NewCMap() *CMap {
	return &CMap{
		Version: 0,
		Nodes:   make([]Node, 0),
	}
}

// SearchCall returns a new search call.
func (m *CMap) SearchCall() *SearchCall {
	return &SearchCall{
		m: m,
		i: ID(-1),
		n: "",
	}
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
