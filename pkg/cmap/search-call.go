package cmap

import (
	"fmt"
)

// SearchCallNode is a handle of search call node operation.
type SearchCallNode struct {
	manager  *cMapManager
	id       ID
	name     NodeName
	nType    NodeType
	status   NodeStatus
	volumeID ID
	random   bool
}

// ID set the node id search condition.
func (c *SearchCallNode) ID(id ID) *SearchCallNode {
	c.id = id
	return c
}

// Name set the node name search condition.
func (c *SearchCallNode) Name(name NodeName) *SearchCallNode {
	c.name = name
	return c
}

// Type set the node type search condition.
func (c *SearchCallNode) Type(t NodeType) *SearchCallNode {
	c.nType = t
	return c
}

// Status set the node status search condition.
func (c *SearchCallNode) Status(s NodeStatus) *SearchCallNode {
	c.status = s
	return c
}

// Random will return the result in random.
func (c *SearchCallNode) Random() *SearchCallNode {
	c.random = true
	return c
}

// Do returns the search result.
func (c *SearchCallNode) Do() (Node, error) {
	m := c.manager.cMaps[c.manager.latest]

	var randIdx []int
	if c.random {
		randIdx = c.manager.random.Perm(len(m.Nodes))
	}

	for i := 0; i < len(m.Nodes); i++ {
		var n Node
		if c.random {
			n = m.Nodes[randIdx[i]]
		} else {
			n = m.Nodes[i]
		}

		// If search condition for ID is set, but the ID is not matched.
		if c.id != ID(-1) && n.ID != c.id {
			continue
		}

		// If search condition for name is set, but the name is not matched.
		if c.name != NodeName("") && n.Name != c.name {
			continue
		}

		// If search condition for type is set, but the type is not matched.
		if c.nType.String() != unknown && n.Type != c.nType {
			continue
		}

		// If search condition for status is set, but the status is not matched.
		if c.status.String() != unknown && n.Stat != c.status {
			continue
		}

		return n, nil
	}

	return Node{}, fmt.Errorf("failed to search the node with the conditions")
}

// SearchCallEncGrp is a handle of search call encoding group operation.
type SearchCallEncGrp struct {
	manager *cMapManager
	id      ID
	status  EncodingGroupStatus
	random  bool
}

// ID set the node id search condition.
func (c *SearchCallEncGrp) ID(id ID) *SearchCallEncGrp {
	c.id = id
	return c
}

// Status set the node status search condition.
func (c *SearchCallEncGrp) Status(s EncodingGroupStatus) *SearchCallEncGrp {
	c.status = s
	return c
}

// Random will return the result in random.
func (c *SearchCallEncGrp) Random() *SearchCallEncGrp {
	c.random = true
	return c
}

// Do returns the search result.
func (c *SearchCallEncGrp) Do() (EncodingGroup, error) {
	m := c.manager.cMaps[c.manager.latest]

	var randIdx []int
	if c.random {
		randIdx = c.manager.random.Perm(len(m.EncGrps))
	}

	for i := 0; i < len(m.EncGrps); i++ {
		var eg EncodingGroup
		if c.random {
			eg = m.EncGrps[randIdx[i]]
		} else {
			eg = m.EncGrps[i]
		}

		if c.id != ID(-1) && c.id != eg.ID {
			continue
		}

		if c.status.String() != unknown && c.status != eg.Stat {
			continue
		}

		return eg, nil
	}

	return EncodingGroup{}, fmt.Errorf("failed to search the encoding group with the conditions")
}

// SearchCallVolume is a handle of search volume call.
type SearchCallVolume struct {
	manager *cMapManager
	id      ID
	status  VolumeStatus
}

// ID set the node id search condition.
func (c *SearchCallVolume) ID(id ID) *SearchCallVolume {
	c.id = id
	return c
}

// Status set the node status search condition.
func (c *SearchCallVolume) Status(s VolumeStatus) *SearchCallVolume {
	c.status = s
	return c
}

// Do returns the search result.
func (c *SearchCallVolume) Do() (Volume, error) {
	m := c.manager.cMaps[c.manager.latest]

	for _, v := range m.Vols {
		if c.id != ID(-1) && c.id != v.ID {
			continue
		}

		if c.status.String() != unknown && c.status != v.Stat {
			continue
		}

		return v, nil
	}

	return Volume{}, fmt.Errorf("failed to search the encoding group with the conditions")
}
