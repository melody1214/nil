package cmap

import (
	"errors"
	"math/rand"
	"time"
)

var (
	// ErrNotFound is returned when failed to search specific components;
	// node, volume, encoding group with the given conditions.
	ErrNotFound = errors.New("no search result with the given conditions")

	// ErrInvalidOptions is returned when some options are set with not proper operation.
	// For example, find max of something with DoAll operation.
	ErrInvalidOptions = errors.New("some options are invalid with this operation")
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

// Version returns the version of searching cluster map.
func (c *SearchCall) Version() Version {
	return c.cmap.Version
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
	cmap    *CMap
	manager *manager
	id      ID
	name    NodeName
	nType   NodeType
	status  NodeStatus
	random  bool
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

	return Node{}, ErrNotFound
}

func init() {
	// Initializes random seed.
	random = rand.New(rand.NewSource(time.Now().Unix()))
}
