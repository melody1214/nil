package gencoding

// Repository provides access to gencoding database.
type Repository interface {
	// Begin() (repository.TxID, error)
	// Rollback(repository.TxID) error
	// Commit(repository.TxID) error
	// AmILeader() (bool, error)
	// LeaderEndpoint() string
	// RegionEndpoint(regionID int) (endpoint string)
}
