package gencoding

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/gencoding/token"
)

// Repository provides access to gencoding database.
type Repository interface {
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	GenerateGencodingGroup(regions []string) error
	UpdateUnencodedChunks(regionName string, unencoded int) error
	AmILeader() bool
	LeaderEndpoint() string
	Make() (*Table, error)
	RegionEndpoint(regionID int) (endpoint string)
	GetRoutes(leaderEndpoint string) (*token.Leg, error)
}
