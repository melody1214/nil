package gencoding

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/gencoding/token"
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides access to gencoding database.
type Repository interface {
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	GenerateGencodingGroup(regions []string) error
	// UpdateUnencodedChunks(regionName string, unencoded int) error
	AmILeader() bool
	LeaderEndpoint() string
	// Make() (*Table, error)
	RegionEndpoint(regionID int) (endpoint string)
	GetRoutes(leaderEndpoint string) (*token.Leg, error)
	MakeGlobalEncodingJob(token *token.Token, primary *token.Unencoded) error
	GetJob(regionName string) (*token.Token, error)
	SetJobStatus(id int64, status Status) error
	RemoveFailedJobs() error
	JobFinished(*token.Token) error
	UpdateUnencoded(egs []cmap.EncodingGroup) ([]cmap.EncodingGroup, error)
	GetChunk(eg cmap.ID) (cID string, err error)
	SetChunk(cID string, egID cmap.ID, status string) error
	GetCandidateChunk(egID cmap.ID, region string) (cID string, err error)
	SetPrimaryChunk(token.Unencoded, int64) error
}
