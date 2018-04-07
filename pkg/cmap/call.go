package cmap

import (
	"fmt"
)

// SearchCall is a handle of search call operation.
type SearchCall struct {
	m *Controller
	i ID
	n string
	t Type
	s Status
}

// ID set the node id search condition.
func (c *SearchCall) ID(id ID) *SearchCall {
	c.i = id
	return c
}

// Name set the node name search condition.
func (c *SearchCall) Name(name string) *SearchCall {
	c.n = name
	return c
}

// Type set the node type search condition.
func (c *SearchCall) Type(t Type) *SearchCall {
	c.t = t
	return c
}

// Status set the node status search condition.
func (c *SearchCall) Status(s Status) *SearchCall {
	c.s = s
	return c
}

// Do returns the search result.
func (c *SearchCall) Do() (Node, error) {
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
