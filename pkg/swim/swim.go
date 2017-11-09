package swim

import "github.com/chanyoung/nil/pkg/swim/swimpb"

// compare compares two member which one is we already has an memlist, and
// the other is new incoming value from the ping message. This returns true
// if old membership information is outdated and need to be updated to new one.
func compare(old, new *swimpb.Member) bool {
	// We don't have the information about new member before.
	if old == nil {
		return true
	}

	// Incoming information is same with ours.
	if old.Incarnation == new.Incarnation && old.Status == new.Status {
		return false
	}

	// Incoming information is same with ours.
	// Our information is more up-to-date.
	if old.Incarnation > new.Incarnation {
		return false
	}

	// Compare status and incartion.
	// See the paper for detailed information.
	switch new.Status {
	case swimpb.Status_ALIVE:
		if old.Status == swimpb.Status_ALIVE && old.Incarnation < new.Incarnation {
			return true
		}

		if old.Status == swimpb.Status_SUSPECT && old.Incarnation < new.Incarnation {
			return true
		}

	case swimpb.Status_SUSPECT:
		if old.Status == swimpb.Status_ALIVE && old.Incarnation <= new.Incarnation {
			return true
		}

		if old.Status == swimpb.Status_SUSPECT && old.Incarnation < new.Incarnation {
			return true
		}

	case swimpb.Status_FAULTY:
		return true
	}

	return false
}
