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

func (ml *memList) set(m *swimpb.Member) {
	ml.Lock()
	defer ml.Unlock()

	if compare(ml.list[m.Uuid], m) {
		ml.list[m.Uuid] = m
	}
}

func (ml *memList) changeStatus(id string, status swimpb.Status) {
	ml.Lock()
	defer ml.Unlock()

	m := ml.list[id]
	if m == nil {
		return
	}

	m.Status = status
	m.Incarnation++
}

// Do not change the contents you got.
// Use member slice to read only purpose.
func (ml *memList) getAll() []*swimpb.Member {
	ml.Lock()
	defer ml.Unlock()

	s := make([]*swimpb.Member, 0)
	for _, m := range ml.list {
		s = append(s, m)
	}

	return s
}
