package clustermap

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cluster"
)

// Repository provides access to cluster map database.
type Repository interface {
	FindAllNodes(txid repository.TxID) ([]cluster.Node, error)
	FindAllVolumes(txid repository.TxID) (vols []cluster.Volume, err error)
	FindAllEncGrps(txid repository.TxID) (EngGrps []cluster.EncodingGroup, err error)
	GetNewClusterMapVer(txid repository.TxID) (cluster.CMapVersion, error)
	JoinNewNode(node cluster.Node) error
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
}
