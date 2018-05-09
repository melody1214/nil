package cluster

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
)

// eventType represents the type of occured event.
type eventType int

const (
	addNode eventType = iota
	fail
)

// eventTime represents the occured time of the event.
type eventTime string

func (t eventTime) String() string { return string(t) }

// event holds the information about what event's are occured.
// This is an value object.
type event struct {
	eType     eventType
	affected  cmap.EncodingGroup
	timeStamp eventTime
}

func newEvent(eType eventType, affected cmap.EncodingGroup) *event {
	return &event{
		eType:     eType,
		affected:  affected,
		timeStamp: eventTime(time.Now().String()),
	}
}
