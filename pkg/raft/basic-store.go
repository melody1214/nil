package raft

import (
	"fmt"
	"sync"
)

type basicStore struct {
	index uint64
	logs  map[uint64]string
	mu    sync.RWMutex
}

func newBasicStore() *basicStore {
	return &basicStore{
		logs: make(map[uint64]string),
	}
}

func (bs *basicStore) lastIndex() uint64 {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.index
}

func (bs *basicStore) readLog(index uint64) (*logEntry, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	if q, ok := bs.logs[index]; ok {
		return &logEntry{
			index: index,
			query: q,
		}, nil
	}

	return nil, fmt.Errorf("no entry with the given index %d", index)
}

func (bs *basicStore) writeLog(logEntry *logEntry) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.logs[logEntry.index] = logEntry.query
	if bs.index < logEntry.index {
		bs.index = logEntry.index
	}

	return nil
}
