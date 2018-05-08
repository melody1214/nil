package cluster

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
)

// Repository provides access to cmap map database.
type Repository interface {
	FindAllNodes(txid repository.TxID) ([]cmap.Node, error)
	FindAllVolumes(txid repository.TxID) (vols []cmap.Volume, err error)
	FindAllEncGrps(txid repository.TxID) (EngGrps []cmap.EncodingGroup, err error)
	GetNewClusterMapVer(txid repository.TxID) (cmap.Version, error)
	LocalJoin(node cmap.Node) error
	GlobalJoin(raftAddr, nodeID string) error
	Execute(txid repository.TxID, query string) (sql.Result, error)
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	Open(raftL *nilmux.Layer) error
	Close() error
}
