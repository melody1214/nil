package db

import (
	"sync"
)

// BasicDB is for testing in-memory key-value store.
type BasicDB struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

// New returns a new BasicDB object.
func New() *BasicDB {
	return &BasicDB{
		data: map[string]interface{}{},
	}
}

// Put inserts value into the BasicDB with the given key.
func (b *BasicDB) Put(key string, val interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.data[key] = val
}

// Get returns value of given key, nil otherwise.
func (b *BasicDB) Get(key string) interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.data[key]
}

// Delete deletes the element with the given key.
func (b *BasicDB) Delete(key string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	delete(b.data, key)
}
