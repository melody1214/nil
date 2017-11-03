package swim

import (
	"sync"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

type memList struct {
	list map[uuid.UUID]*member
	mu   sync.RWMutex
}

func newMemList() *memList {
	return &memList{
		list: make(map[uuid.UUID]*member),
	}
}
