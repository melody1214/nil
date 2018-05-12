package cluster

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
)

// EventType represents the type of occured event.
type EventType int

const (
	// LocalJoin means the event type is local join.
	LocalJoin EventType = iota
	// RegisterVolume means the event type is register new volume.
	RegisterVolume
	// Fail means the event type is fail.
	Fail
)

// String returns the string type of its value.
func (t EventType) String() string {
	if t == LocalJoin {
		return "LocalJoin"
	} else if t == RegisterVolume {
		return "RegisterVolume"
	} else if t == Fail {
		return "Fail"
	}
	return "Unknown"
}

// Time represents the occured time of the event or job.
type Time string

// String returns its value in built-in string type.
func (t Time) String() string { return string(t) }

// TimeNow returns the current time with format.
func TimeNow() Time {
	return Time(time.Now().Format(time.RFC3339))
}

// Event holds the information about what event's are occured.
// This is an value object.
type Event struct {
	Type       EventType
	AffectedEG cmap.ID
	TimeStamp  Time
}

// NoAffectedEG means this event has not affect to any encoding group.
const NoAffectedEG = cmap.ID(-1)

func newEvent(t EventType, affectedEG cmap.ID) *Event {
	return &Event{
		Type:       t,
		AffectedEG: affectedEG,
		TimeStamp:  TimeNow(),
	}
}
