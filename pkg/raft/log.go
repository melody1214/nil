package raft

// logEntry is the entry of the raft logs.
type logEntry struct {
	index uint64
	query string
}

// logStore is the store of the raft logs.
type logStore interface {
	// lastIndex returns the last applied index.
	lastIndex() (index uint64)

	// readLog returns a log entry with the given index.
	readLog(index uint64) (logEntry *logEntry, err error)

	// writeLog write a log entry into the store.
	writeLog(logEntry *logEntry) error
}
