package object

// Repository provides access to object database.
type Repository interface {
	// Put(o *ObjInfo) error
	// Get(name string) (*ObjInfo, error)
	// GetChunk(eg cmap.ID) (cID string, err error)
	// SetChunk(cID string, egID cmap.ID, status string) error
}
