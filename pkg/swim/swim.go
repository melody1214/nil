package swim

// compare compares two member which one is we already has an memlist, and
// the other is new incoming value from the ping message. This returns true
// if old membership information is outdated and need to be updated to new one.
func compare(old, new *Member) bool {
	// We don't have the information about new member before.
	if old == nil {
		return true
	}

	// Compare status and incartion.
	// See the paper for detailed information.
	switch new.Status {
	case Alive:
		if old.Status == Alive && old.Incarnation < new.Incarnation {
			return true
		}

		if old.Status == Suspect && old.Incarnation < new.Incarnation {
			return true
		}

	case Suspect:
		if old.Status == Alive && old.Incarnation <= new.Incarnation {
			return true
		}

		if old.Status == Suspect && old.Incarnation < new.Incarnation {
			return true
		}

	case Faulty:
		return true
	}

	return false
}
