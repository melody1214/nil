package swim

import (
	"sync"
)

type memList struct {
	list map[string]*Member
	sync.Mutex
}

func newMemList() *memList {
	return &memList{
		list: make(map[string]*Member),
	}
}

// get returns copied object of the given id.
func (ml *memList) get(id string) *Member {
	ml.Lock()
	defer ml.Unlock()

	m := ml.list[id]
	if m == nil {
		return m
	}

	cp := &Member{}
	*cp = *m
	return cp
}

// set compares the given member object is newer than mine.
// If newer than mine, then update it.
func (ml *memList) set(m *Member) {
	ml.Lock()
	defer ml.Unlock()

	if compare(ml.list[m.UUID], m) {
		ml.list[m.UUID] = m
	}
}

// Fetch fetches 'n' random members from the member list.
// Fetch all items if n <= 0.
func (ml *memList) fetch(n int) []*Member {
	ml.Lock()
	defer ml.Unlock()

	if n <= 0 {
		n = len(ml.list)
	}

	fetched := make([]*Member, n)
	for _, v := range ml.list {
		if n--; n < 0 {
			break
		}

		cv := &Member{}
		*cv = *v
		fetched[n] = cv
	}

	return fetched
}
