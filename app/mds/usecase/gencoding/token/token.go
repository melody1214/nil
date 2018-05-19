package token

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
)

// Token contains all the information for a single global encoding operation.
// The token traverses each region by routing information and collects about
// the unencoded candidate chunks in each region. Each region can put a
// candidate chunk into the token according to its priority, and the token
// will eventually return to the master node of the global cluster.
type Token struct {
	Version int
	Timeout time.Time
	Routing Leg
	First   Unencoded
	Second  Unencoded
	Third   Unencoded
	Primary Unencoded
}

// Add adds the new unencoded token if the priority is higher than old one.
func (t *Token) Add(new Unencoded) {
	candidates := [3]*Unencoded{&t.First, &t.Second, &t.Third}
	for _, c := range candidates {
		if new.Priority <= c.Priority {
			continue
		}
		*c = new
		break
	}
}

// Endpoint represents the endpoint of the region.
type Endpoint string

// Leg is the collection of routing stops.
type Leg struct {
	CurrentIdx int
	Stops      []Stop
}

// NewLeg creates the new leg with the given stops.
func NewLeg(stops ...Stop) *Leg {
	return &Leg{
		CurrentIdx: 0,
		Stops:      stops,
	}
}

// Next returns the next routin point.
func (l *Leg) Next() Stop {
	if l.Stops == nil {
		return Stop{}
	}

	l.CurrentIdx = l.CurrentIdx + 1

	if l.CurrentIdx >= len(l.Stops) {
		return l.Stops[0]
	}

	return l.Stops[l.CurrentIdx]
}

// Current returns the current stop.
func (l *Leg) Current() Stop {
	if l.CurrentIdx >= len(l.Stops) {
		return l.Stops[0]
	}
	return l.Stops[l.CurrentIdx]
}

// Stop represents the specific region information for routing.
type Stop struct {
	RegionID   int64
	RegionName string
	Endpoint   Endpoint
}

// Unencoded contains the global encoding candidate chunk information.
type Unencoded struct {
	Region   Stop
	Node     cmap.ID
	Volume   cmap.ID
	EncGrp   cmap.ID
	ChunkID  string
	Priority int
}
