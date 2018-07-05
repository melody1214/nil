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
	// Rebalance means the event type is rebalance.
	Rebalance
	// Update means need to update cluster map from db data.
	Update
	// Fail means the event type is fail.
	Fail
)

// String returns the string type of its value.
func (t EventType) String() string {
	if t == LocalJoin {
		return "LocalJoin"
	} else if t == RegisterVolume {
		return "RegisterVolume"
	} else if t == Rebalance {
		return "Rebalance"
	} else if t == Fail {
		return "Fail"
	} else if t == Update {
		return "Update"
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

func extractEventsFromCMap(old, new *cmap.CMap) []*Event {
	needRebalance := false
	needUpdate := false

	events := make([]*Event, 0)
	for _, newNode := range new.Nodes {
		for _, oldNode := range old.Nodes {
			if newNode.ID != oldNode.ID {
				continue
			}

			if newNode.Stat == oldNode.Stat {
				continue
			}

			switch newNode.Stat {
			case cmap.NodeSuspect:
				needUpdate = true
				// Make nodes rdonly.
			case cmap.NodeFaulty:
				// Make recovery events.
				for _, vID := range newNode.Vols {
					for _, v := range new.Vols {
						if v.ID != vID {
							continue
						}
						for _, egID := range v.EncGrps {
							e := newEvent(Fail, egID)
							events = append(events, e)
						}
						break
					}
				}
			}
		}
	}

	for _, newVol := range new.Vols {
		for _, oldVol := range old.Vols {
			if newVol.ID != oldVol.ID {
				continue
			}

			if newVol.Stat == oldVol.Stat {
				continue
			}

			switch oldVol.Stat {
			case cmap.VolPrepared:
				if newVol.Stat == cmap.VolActive {
					needRebalance = true
				} else if newVol.Stat == cmap.VolFailed {
					// Make recovery events.
				}
			case cmap.VolActive:
				// Make recovery events.
			}
		}
	}

	if needUpdate {
		e := newEvent(Update, NoAffectedEG)
		events = append(events, e)
	}
	if needRebalance {
		e := newEvent(Rebalance, NoAffectedEG)
		events = append(events, e)
	}

	return events
}
