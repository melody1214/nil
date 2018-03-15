package cmap

import (
	"fmt"
)

// CMap is a cluster map which includes the information about nodes.
type CMap struct {
	Version int64  `xml:"version"`
	Nodes   []Node `xml:"node"`
}

// SearchCall returns a new search call.
func (m *CMap) SearchCall() *SearchCall {
	return &SearchCall{m: m}
}

// HumanReadable returns a human readable map of the cluster.
func (m *CMap) HumanReadable() string {
	out := ""

	// Make human readable sentences for each nodes.
	for _, n := range m.Nodes {
		row := fmt.Sprintf(
			"| %4s | %s | %7s |\n",
			n.Type.String(),
			n.Addr,
			n.Stat.String(),
		)

		out += row
	}

	return out
}
