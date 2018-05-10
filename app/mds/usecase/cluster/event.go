package cluster

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
)

// EventType represents the type of occured event.
type EventType int

const (
	// AddNode means the event type is add node.
	AddNode EventType = iota
	// Fail means the event type is fail.
	Fail
)

// Time represents the occured time of the event or job.
type Time string

// String returns its value in built-in string type.
func (t Time) String() string { return string(t) }

// Event holds the information about what event's are occured.
// This is an value object.
type Event struct {
	Type       EventType
	AffectedEG cmap.ID
	TimeStamp  Time
}

func newEvent(t EventType, affectedEG cmap.ID) *Event {
	return &Event{
		Type:       t,
		AffectedEG: affectedEG,
		TimeStamp:  Time(time.Now().String()),
	}
}
