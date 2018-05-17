package repository

// Service is a backend store interface.
type Service interface {
	// Run starts to service backend store.
	Run()

	// Stop clean up the backend store and make it stop gracefully.
	Stop()

	// AddVolume adds a new volume into the store list.
	AddVolume(v *Vol) error

	// Push pushes an io request into the scheduling queue.
	Push(c *Request) error

	GetObjectSize(lvID, objID string) (int64, bool)
	GetObjectMD5(lvID, objID string) (string, bool)
	GetChunkHeaderSize() int64
	GetObjectHeaderSize() int64

	RenameChunk(src string, dest string, Vol string, LocGid string) error
	CountNonCodedChunk(Vol string, LocGid string) (int, error)
	GetNonCodedChunk(Vol string, LocGid string) (string, error)
}
