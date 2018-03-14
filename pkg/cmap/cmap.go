package cmap

import (
	"fmt"
)

// GetLatest get the latest cluster map.
func GetLatest(mdsAddrs ...string) (*CMap, error) {
	if len(mdsAddrs) == 0 {
		mdsAddr, err := getMdsAddr()
		if err != nil {
			return nil, fmt.Errorf("TODO: read mds address from file")
		}

		mdsAddrs = append(mdsAddrs, mdsAddr)
	}

	for _, mdsAddr := range mdsAddrs {
		return getLatest(mdsAddr)
	}

	return nil, fmt.Errorf("couldn't get the mds address")
}

func getLatest(mdsAddr string) (*CMap, error) {
	return nil, nil
}

func getMdsAddr() (string, error) {
	return "", fmt.Errorf("not implemented")
}

// CMap is a cluster map which includes the information about nodes.
type CMap struct {
	OutDated bool   `xml:"outdated"`
	Nodes    []Node `xml:"node"`
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
