package swim

import (
	"sync"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

type memList struct {
	list map[uuid.UUID]*member
	sync.Mutex
}

func newMemList() *memList {
	return &memList{
		list: make(map[uuid.UUID]*member),
	}
}

func (ml *memList) set(m *member) {
	ml.Lock()
	defer ml.Unlock()

	old := ml.list[m.id]
	if old == nil && m.status != FAULTY {
		ml.list[m.id] = m
		return
	}

	switch m.status {
	case ALIVE:
		if old.status == ALIVE && old.incarnation < m.incarnation {
			ml.list[m.id] = m
			return
		}

		if old.status == SUSPECT && old.incarnation < m.incarnation {
			ml.list[m.id] = m
			return
		}

	case SUSPECT:
		if old.status == ALIVE && old.incarnation <= m.incarnation {
			ml.list[m.id] = m
			return
		}

		if old.status == SUSPECT && old.incarnation < m.incarnation {
			ml.list[m.id] = m
			return
		}

	case FAULTY:
		delete(ml.list, m.id)
	}
}
