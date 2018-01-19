package swim

import (
	"sync"
	"syscall"
)

type memList struct {
	myID ServerID

	list map[ServerID]Member
	sync.Mutex
}

func newMemList(myID ServerID) *memList {
	return &memList{
		myID: myID,
		list: make(map[ServerID]Member),
	}
}

// get returns member structure of the given id.
func (ml *memList) get(id ServerID) (m Member, ok bool) {
	ml.Lock()
	defer ml.Unlock()

	m, ok = ml.list[id]
	return
}

// set compares the given member object is newer than mine.
// If newer than mine, then update it.
func (ml *memList) set(new Member) {
	ml.Lock()
	defer ml.Unlock()

	old, ok := ml.list[new.ID]
	// New member always add into the member list.
	if !ok {
		ml.list[new.ID] = new
	}

	// Update information about myself.
	if ml.myID == new.ID {
		// If some other nodes tell I am suspect state,
		if new.Status == Suspect && old.Incarnation <= new.Incarnation {
			old.Status = Alive
			old.Incarnation++
			ml.list[new.ID] = old
		}

		// I'm faulty node.
		// Must stop the system even if I'm alive.
		if new.Status == Faulty {
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}

	if compare(old, new) {
		ml.list[new.ID] = new
	}
}

// Fetch fetches 'n' random members from the member list.
// Fetch all items if n < 1.
func (ml *memList) fetch(n int, opts ...fetchOption) []Member {
	ml.Lock()
	defer ml.Unlock()

	var fopts fetchOptions
	for _, opt := range opts {
		opt(&fopts)
	}

	if n < 1 {
		n = len(ml.list)
	}

	fetched := make([]Member, 0, n)
	for _, m := range ml.list {
		if fopts.notAlive && m.Status == Alive {
			continue
		}
		if fopts.notSuspect && m.Status == Suspect {
			continue
		}
		if fopts.notFaulty && m.Status == Faulty {
			continue
		}
		if fopts.notMyself && m.ID == ml.myID {
			continue
		}

		fetched = append(fetched, m)
		if len(fetched) == n {
			break
		}
	}

	return fetched
}

// fetchOptions configure a fetch call.
type fetchOptions struct {
	notAlive   bool
	notSuspect bool
	notFaulty  bool
	notMyself  bool
}

// fetchOption configures how we fetch the members.
type fetchOption func(*fetchOptions)

// withNotAlive returns a fetchOption which fetches except alive members.
func withNotAlive() fetchOption {
	return func(o *fetchOptions) {
		o.notAlive = true
	}
}

// withNotSuspect returns a fetchOption which fetches except suspect members.
func withNotSuspect() fetchOption {
	return func(o *fetchOptions) {
		o.notSuspect = true
	}
}

// withNotFaulty returns a fetchOption which fetches except faulty members.
func withNotFaulty() fetchOption {
	return func(o *fetchOptions) {
		o.notFaulty = true
	}
}

// withNotMyself returns a fetchOption which fetches except me.
func withNotMyself() fetchOption {
	return func(o *fetchOptions) {
		o.notMyself = true
	}
}
