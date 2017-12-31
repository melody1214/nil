package swim

import (
	"time"

	"github.com/chanyoung/nil/pkg/util/config"
)

var (
	// Time interval of generates ping message.
	// Swim server will sends ping periodically with this interval.
	pingPeriod = 3 * time.Second

	// Expire time of ping messages.
	pingExpire = 3 * time.Second
)

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

func init() {
	pp, e := time.ParseDuration(config.Get("swim.period"))
	if e == nil {
		pingPeriod = pp
	}
	pe, e := time.ParseDuration(config.Get("swim.expire"))
	if e == nil {
		pingExpire = pe
	}
}
