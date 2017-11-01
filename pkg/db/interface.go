package db

// DB interface wraps the operations of a key/value store.
type DB interface {
	// Put inserts value into the KV store with the given key.
	// To update the value of existing key, put a value with same key.
	Put(string, interface{})

	// Get retrieves value of given key.
	// It returns nil if no matching key.
	Get(string) interface{}

	// Delete removes the element from the db.
	Delete(key string)
}
