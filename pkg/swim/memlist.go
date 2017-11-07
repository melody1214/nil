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

	old := ml.list[m.Uuid]
	if old == nil && m.Status != swimpb.Status_FAULTY {
		ml.list[m.Uuid] = m
		return
	}

	switch m.Status {
	case swimpb.Status_ALIVE:
		if old.Status == swimpb.Status_ALIVE && old.Incarnation < m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

		if old.Status == swimpb.Status_SUSPECT && old.Incarnation < m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

	case swimpb.Status_SUSPECT:
		if old.Status == swimpb.Status_ALIVE && old.Incarnation <= m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

		if old.Status == swimpb.Status_SUSPECT && old.Incarnation < m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

	case swimpb.Status_FAULTY:
		delete(ml.list, m.Uuid)
	}
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
