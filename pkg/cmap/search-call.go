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

// EncGrp returns a SearchCallEncGrp object for searching encoding group member.
func (c *SearchCall) EncGrp() *SearchCallEncGrp {
	return &SearchCallEncGrp{
		cmap:      c.cmap,
		id:        ID(-1),
		leaderVol: ID(-1),
		status:    EncodingGroupStatus(unknown),
		random:    false,
		minUenc:   false,
		maxUenc:   false,
	}
}

// SearchCallEncGrp is a handle of search call encoding group operation.
type SearchCallEncGrp struct {
	cmap      *CMap
	manager   *manager
	id        ID
	leaderVol ID
	status    EncodingGroupStatus
	random    bool
	minUenc   bool
	maxUenc   bool
}

// ID set the node id search condition.
func (c *SearchCallEncGrp) ID(id ID) *SearchCallEncGrp {
	c.id = id
	return c
}

// LeaderVol set the leader volume id search condition.
func (c *SearchCallEncGrp) LeaderVol(id ID) *SearchCallEncGrp {
	c.leaderVol = id
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

// MinUenc will return the result who has the minimum number of unencoded chunks.
func (c *SearchCallEncGrp) MinUenc() *SearchCallEncGrp {
	c.minUenc = true
	return c
}

// MaxUenc will return the result who has the maximum number of unencoded chunks.
func (c *SearchCallEncGrp) MaxUenc() *SearchCallEncGrp {
	c.maxUenc = true
	return c
}

// Do returns the search result.
func (c *SearchCallEncGrp) Do() (EncodingGroup, error) {
	var result = EncodingGroup{
		ID:   ID(0),
		Uenc: 0,
	}
	var randIdx []int

	// Setup random index if call option is set random.
	if c.random {
		randIdx = random.Perm(len(c.cmap.EncGrps))
	}

	// Lookup all encoding groups.
	for i := 0; i < len(c.cmap.EncGrps); i++ {
		eg := c.cmap.EncGrps[i]
		if c.random {
			eg = c.cmap.EncGrps[randIdx[i]]
		}

		if c.id != ID(-1) && c.id != eg.ID {
			continue
		}

		if c.leaderVol != ID(-1) {
			if len(eg.Vols) == 0 {
				continue
			}

			if c.leaderVol != eg.LeaderVol() {
				continue
			}
		}

		if c.status.String() != unknown && c.status != eg.Stat {
			continue
		}

		if c.maxUenc && result.Uenc <= eg.Uenc {
			result = eg
			continue
		}

		if c.minUenc && result.Uenc >= eg.Uenc {
			result = eg
			continue
		}

		result = eg
		break
	}

	// Failed to find.
	if result.ID == ID(0) {
		return result, ErrNotFound
	}
	return result, nil
}

// DoAll can returns multiple search results.
func (c *SearchCallEncGrp) DoAll() ([]EncodingGroup, error) {
	result := make([]EncodingGroup, 0)
	var randIdx []int

	// These options are not valid with DoAll operation.
	if c.maxUenc || c.minUenc || c.id != ID(-1) {
		return result, ErrInvalidOptions
	}

	// Setup random index if call option is set random.
	if c.random {
		randIdx = random.Perm(len(c.cmap.EncGrps))
	}

	// Lookup all encoding groups.
	for i := 0; i < len(c.cmap.EncGrps); i++ {
		eg := c.cmap.EncGrps[i]
		if c.random {
			eg = c.cmap.EncGrps[randIdx[i]]
		}

		if c.leaderVol != ID(-1) {
			if len(eg.Vols) == 0 {
				continue
			}

			if c.leaderVol != eg.LeaderVol() {
				continue
			}
		}

		if c.status.String() != unknown && c.status != eg.Stat {
			continue
		}

		result = append(result, eg)
	}

	// Failed to find.
	if len(result) == 0 {
		return result, ErrNotFound
	}
	return result, nil
}

// Volume returns a SearchCallVolume object for searching volume member.
func (c *SearchCall) Volume() *SearchCallVolume {
	return &SearchCallVolume{
		cmap:   c.cmap,
		id:     ID(-1),
		node:   ID(-1),
		status: VolumeStatus(unknown),
	}
}

// SearchCallVolume is a handle of search volume call.
type SearchCallVolume struct {
	cmap    *CMap
	manager *manager
	id      ID
	node    ID
	status  VolumeStatus
}

// ID set the volume id search condition.
func (c *SearchCallVolume) ID(id ID) *SearchCallVolume {
	c.id = id
	return c
}

// Node set the node id search condition.
func (c *SearchCallVolume) Node(id ID) *SearchCallVolume {
	c.node = id
	return c
}

// Status set the node status search condition.
func (c *SearchCallVolume) Status(s VolumeStatus) *SearchCallVolume {
	c.status = s
	return c
}

// Do returns the search result.
func (c *SearchCallVolume) Do() (Volume, error) {
	for _, v := range c.cmap.Vols {
		if c.id != ID(-1) && c.id != v.ID {
			continue
		}

		if c.node != ID(-1) && c.node != v.Node {
			continue
		}

		if c.status.String() != unknown && c.status != v.Stat {
			continue
		}

		return v, nil
	}

	return Volume{}, ErrNotFound
}

// DoAll can returns multiple search results.
func (c *SearchCallVolume) DoAll() ([]Volume, error) {
	result := make([]Volume, 0)

	// These options are not valid with DoAll operation.
	if c.id != ID(-1) {
		return result, ErrInvalidOptions
	}

	// Lookup all encoding groups.
	for i := 0; i < len(c.cmap.Vols); i++ {
		v := c.cmap.Vols[i]

		if c.node != ID(-1) && c.node != v.Node {
			continue
		}

		if c.status.String() != unknown && c.status != v.Stat {
			continue
		}

		result = append(result, v)
	}

	// Failed to find.
	if len(result) == 0 {
		return result, ErrNotFound
	}
	return result, nil
}

func init() {
	// Initializes random seed.
	random = rand.New(rand.NewSource(time.Now().Unix()))
}
