package cluster

// Repository provides access to cmap map database.
type Repository interface {
	// FindAllNodes(repository.TxID) ([]cmap.Node, error)
	// GetNewClusterMapVer(repository.TxID) (cmap.Version, error)
	// LocalJoin(cmap.Node) error
	// GlobalJoin(raftAddr, nodeID string) error

	// Begin() (repository.TxID, error)
	// Rollback(repository.TxID) error
	// Commit(repository.TxID) error
	// Open(raftL *nilmux.Layer) error
	// Close() error
}
