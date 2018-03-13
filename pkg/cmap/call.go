package cmap

// SearchCall is a handle of search call operation.
type SearchCall struct {
	m *CMap
	t Type
	s Status
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
func (c *SearchCall) Do() *Node {
	return &Node{}
}
