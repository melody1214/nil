package clustermap

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides access to cluster map database.
type Repository interface {
	FindAllNodes(txid repository.TxID) ([]cmap.Node, error)
	GetNewClusterMapVer(txid repository.TxID) (cmap.Version, error)
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
}
