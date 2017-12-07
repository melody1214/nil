package raft

// logs store.
type log struct {
	index uint64
	data  []string
}
