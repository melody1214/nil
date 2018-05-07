package clustermap

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides access to cmap map database.
type Repository interface {
	FindAllNodes(txid repository.TxID) ([]cmap.Node, error)
	FindAllVolumes(txid repository.TxID) (vols []cmap.Volume, err error)
	FindAllEncGrps(txid repository.TxID) (EngGrps []cmap.EncodingGroup, err error)
	GetNewClusterMapVer(txid repository.TxID) (cmap.Version, error)
	JoinNewNode(node cmap.Node) error
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
}
