package swim

import "sync"

type memList struct {
	list map[string]*Member
	sync.Mutex
}

func newMemList() *memList {
	return &memList{
		list: make(map[string]*Member),
	}
}

func (ml *memList) set(m *Member) {
	ml.Lock()
	defer ml.Unlock()

	old := ml.list[m.Uuid]
	if old == nil && m.Status != Status_FAULTY {
		ml.list[m.Uuid] = m
		return
	}

	switch m.Status {
	case Status_ALIVE:
		if old.Status == Status_ALIVE && old.Incarnation < m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

		if old.Status == Status_SUSPECT && old.Incarnation < m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

	case Status_SUSPECT:
		if old.Status == Status_ALIVE && old.Incarnation <= m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

		if old.Status == Status_SUSPECT && old.Incarnation < m.Incarnation {
			ml.list[m.Uuid] = m
			return
		}

	case Status_FAULTY:
		delete(ml.list, m.Uuid)
	}
}
