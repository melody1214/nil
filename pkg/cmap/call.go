package cmap

import (
	"fmt"
)

// SearchCallNode is a handle of search call node operation.
type SearchCallNode struct {
	m   *Controller
	i   ID
	n   string
	t   Type
	s   Status
	vid ID
}

// ID set the node id search condition.
func (c *SearchCallNode) ID(id ID) *SearchCallNode {
	c.i = id
	return c
}

// Name set the node name search condition.
func (c *SearchCallNode) Name(name string) *SearchCallNode {
	c.n = name
	return c
}

// Type set the node type search condition.
func (c *SearchCallNode) Type(t Type) *SearchCallNode {
	c.t = t
	return c
}

// Status set the node status search condition.
func (c *SearchCallNode) Status(s Status) *SearchCallNode {
	c.s = s
	return c
}

// Do returns the search result.
func (c *SearchCallNode) Do() (Node, error) {
	m := c.m.cMaps[c.m.latest]
	for _, n := range m.Nodes {
		// If search condition for ID is set, but the ID is not matched.
		if c.i != ID(-1) && n.ID != c.i {
			continue
		}

		// If search condition for name is set, but the name is not matched.
		if c.n != "" && n.Name != c.n {
			continue
		}

		// If search condition for type is set, but the type is not matched.
		if c.t.String() != unknown && n.Type != c.t {
			continue
		}

		// If search condition for status is set, but the status is not matched.
		if c.s.String() != unknown && n.Stat != c.s {
			continue
		}

		// All conditions are mathced, return the find node.
		return n, nil
	}

	return Node{}, fmt.Errorf("failed to search the node with the conditions")
}

// SearchCallEncGrp is a handle of search call encoding group operation.
type SearchCallEncGrp struct {
	m *Controller
	i ID
	s EncodingGroupStatus
	r bool
}

// ID set the node id search condition.
func (c *SearchCallEncGrp) ID(id ID) *SearchCallEncGrp {
	c.i = id
	return c
}

// Status set the node status search condition.
func (c *SearchCallEncGrp) Status(s EncodingGroupStatus) *SearchCallEncGrp {
	c.s = s
	return c
}

// Random will return the result in random.
func (c *SearchCallEncGrp) Random() *SearchCallEncGrp {
	c.r = true
	return c
}

// Do returns the search result.
func (c *SearchCallEncGrp) Do() (EncodingGroup, error) {
	m := c.m.cMaps[c.m.latest]

	var randIdx []int
	if c.r {
		randIdx = c.m.random.Perm(len(m.EncGrps))
	}

	for i := 0; i < len(m.EncGrps); i++ {
		var eg EncodingGroup
		if c.r {
			eg = m.EncGrps[randIdx[i]]
		} else {
			eg = m.EncGrps[i]
		}

		if c.i != ID(-1) && c.i != eg.ID {
			continue
		}

		if c.s.String() != unknown && c.s != eg.Status {
			continue
		}

		return eg, nil
	}

	return EncodingGroup{}, fmt.Errorf("failed to search the encoding group with the conditions")
}

// SearchCallVolume is a handle of search volume call.
type SearchCallVolume struct {
	m *Controller
	i ID
	s VolumeStatus
}

// ID set the node id search condition.
func (c *SearchCallVolume) ID(id ID) *SearchCallVolume {
	c.i = id
	return c
}

// Status set the node status search condition.
func (c *SearchCallVolume) Status(s VolumeStatus) *SearchCallVolume {
	c.s = s
	return c
}

// Do returns the search result.
func (c *SearchCallVolume) Do() (Volume, error) {
	m := c.m.cMaps[c.m.latest]

	for _, v := range m.Vols {
		if c.i != ID(-1) && c.i != v.ID {
			continue
		}

		if c.s.String() != unknown && c.s != v.Status {
			continue
		}

		return v, nil
	}

	return Volume{}, fmt.Errorf("failed to search the encoding group with the conditions")
}
