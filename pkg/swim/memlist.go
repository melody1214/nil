package swim

import (
	"sync"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

type memList struct {
	list map[string]*swimpb.Member
	sync.Mutex
}

func newMemList() *memList {
	return &memList{
		list: make(map[string]*swimpb.Member),
	}
}

// get returns copied object of the given id.
func (ml *memList) get(id string) *swimpb.Member {
	ml.Lock()
	defer ml.Unlock()

	m := ml.list[id]
	if m == nil {
		return m
	}

	cp := &swimpb.Member{}
	*cp = *m
	return cp
}

// set compares the given member object is newer than mine.
// If newer than mine, then update it.
func (ml *memList) set(m *swimpb.Member) {
	ml.Lock()
	defer ml.Unlock()

	if compare(ml.list[m.Uuid], m) {
		ml.list[m.Uuid] = m
	}
}

// Fetch fetches 'n' random members from the member list.
// Fetch all items if n <= 0.
func (ml *memList) fetch(n int) []*swimpb.Member {
	ml.Lock()
	defer ml.Unlock()

	if n <= 0 {
		n = len(ml.list)
	}

	fetched := make([]*swimpb.Member, n)
	for _, v := range ml.list {
		if n--; n < 0 {
			break
		}

		cv := &swimpb.Member{}
		*cv = *v
		fetched[n] = cv
	}

	return fetched
}
