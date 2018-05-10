package cluster

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
)

// Repository provides access to cmap map database.
type Repository interface {
	FindAllNodes(repository.TxID) ([]cmap.Node, error)
	FindAllVolumes(repository.TxID) (vols []cmap.Volume, err error)
	FindAllEncGrps(repository.TxID) (EngGrps []cmap.EncodingGroup, err error)
	GetNewClusterMapVer(repository.TxID) (cmap.Version, error)
	LocalJoin(cmap.Node) error
	GlobalJoin(raftAddr, nodeID string) error
	Execute(txid repository.TxID, query string) (sql.Result, error)
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	Open(raftL *nilmux.Layer) error
	Close() error

	// jobRepository methods.
	InsertJob(repository.TxID, *Job) error
}

// jobRepository is repository for storing and tracking jobs.
type jobRepository interface {
	InsertJob(repository.TxID, *Job) error
	// FetchJob() *job
}

func newJobRepository(r Repository) jobRepository {
	return r
}
