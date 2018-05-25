package cmap

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var (
	// ErrNotFound is returned when failed to search specific components;
	// node, volume, encoding group with the given conditions.
	ErrNotFound = errors.New("no search result with the given conditions")
)

// Random for pick random member option.
var random *rand.Rand

// SearchCall is an object that can provide users various and detailed
// search functions.
type SearchCall struct {
	// Cluster map which will be used to search.
	// It is copied through the manager at the time the SearchCall object
	// is created. Because the SearchCall sub methods are not cause changes
	// to the cmap, caller can reuse any number of different sub calls
	// without locking.
	cmap *CMap
}

// Node returns a SearchCallNode object for searching node member.
func (c *SearchCall) Node() *SearchCallNode {
	return &SearchCallNode{
		cmap:   c.cmap,
		id:     ID(-1),
		name:   NodeName(""),
		nType:  NodeType(-1),
		status: NodeStatus(unknown),
	}
}

// SearchCallNode is a handle of search call node operation.
type SearchCallNode struct {
	cmap     *CMap
	manager  *manager
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
	if c.cmap == nil {
		c.cmap = c.manager.cMaps[c.manager.latest]
	}

	var randIdx []int
	if c.random {
		randIdx = random.Perm(len(c.cmap.Nodes))
	}

	for i := 0; i < len(c.cmap.Nodes); i++ {
		var n Node
		if c.random {
			n = c.cmap.Nodes[randIdx[i]]
		} else {
			n = c.cmap.Nodes[i]
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

// EncGrp returns a SearchCallEncGrp object for searching encoding group member.
func (c *SearchCall) EncGrp() *SearchCallEncGrp {
	return &SearchCallEncGrp{
		cmap:   c.cmap,
		id:     ID(-1),
		status: EncodingGroupStatus(unknown),
		random: false,
	}
}

// SearchCallEncGrp is a handle of search call encoding group operation.
type SearchCallEncGrp struct {
	cmap    *CMap
	manager *manager
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
	if c.cmap == nil {
		c.cmap = c.manager.cMaps[c.manager.latest]
	}

	var randIdx []int
	if c.random {
		randIdx = random.Perm(len(c.cmap.EncGrps))
	}

	for i := 0; i < len(c.cmap.EncGrps); i++ {
		var eg EncodingGroup
		if c.random {
			eg = c.cmap.EncGrps[randIdx[i]]
		} else {
			eg = c.cmap.EncGrps[i]
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

// Volume returns a SearchCallVolume object for searching volume member.
func (c *SearchCall) Volume() *SearchCallVolume {
	return &SearchCallVolume{
		cmap:   c.cmap,
		id:     ID(-1),
		status: VolumeStatus(unknown),
	}
}

// SearchCallVolume is a handle of search volume call.
type SearchCallVolume struct {
	cmap    *CMap
	manager *manager
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
	if c.cmap == nil {
		c.cmap = c.manager.cMaps[c.manager.latest]
	}

	for _, v := range c.cmap.Vols {
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

func init() {
	// Initializes random seed.
	random = rand.New(rand.NewSource(time.Now().Unix()))
}
