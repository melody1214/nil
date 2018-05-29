package cluster

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
)

// Repository provides access to cmap map database.
type Repository interface {
	ListJob() []string
	FindAllNodes(repository.TxID) ([]cmap.Node, error)
	FindAllVolumes(repository.TxID) (vols []cmap.Volume, err error)
	FindAllEncGrps(repository.TxID) (EngGrps []cmap.EncodingGroup, err error)
	GetNewClusterMapVer(repository.TxID) (cmap.Version, error)
	LocalJoin(cmap.Node) error
	GlobalJoin(raftAddr, nodeID string) error
	UpdateChangedCMap(txid repository.TxID, cmap *cmap.CMap) error
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	Open(raftL *nilmux.Layer) error
	Close() error

	// jobRepository methods.
	InsertJob(repository.TxID, *Job) error
	FetchJob(txid repository.TxID) (*Job, error)
	UpdateJob(repository.TxID, *Job) error
	RegisterVolume(txid repository.TxID, v *cmap.Volume) error
	MakeNewEncodingGroup(txid repository.TxID, encGrp *cmap.EncodingGroup) error
	FindReplaceableVolume(txid repository.TxID, failedEG *cmap.EncodingGroup, failedVol *cmap.Volume, failureDomain ...cmap.ID) (cmap.ID, error)
	SetEGV(txid repository.TxID, egID cmap.ID, role int, volID, moveTo cmap.ID) error
	VolEGIncr(txid repository.TxID, vID cmap.ID) error
	VolEGDecr(txid repository.TxID, vID cmap.ID) error
	FindAllChunks(egID cmap.ID, status string) ([]int, error)
}

// jobRepository is repository for storing and tracking jobs.
type jobRepository interface {
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	InsertJob(repository.TxID, *Job) error
	FetchJob(txid repository.TxID) (*Job, error)
	UpdateJob(repository.TxID, *Job) error
	LocalJoin(cmap.Node) error

	GetNewClusterMapVer(repository.TxID) (cmap.Version, error)
	FindAllNodes(repository.TxID) ([]cmap.Node, error)
	FindAllVolumes(repository.TxID) (vols []cmap.Volume, err error)
	FindAllEncGrps(repository.TxID) (EngGrps []cmap.EncodingGroup, err error)
	RegisterVolume(txid repository.TxID, v *cmap.Volume) error
	MakeNewEncodingGroup(txid repository.TxID, encGrp *cmap.EncodingGroup) error
	FindReplaceableVolume(txid repository.TxID, failedEG *cmap.EncodingGroup, failedVol *cmap.Volume, failureDomain ...cmap.ID) (cmap.ID, error)
	SetEGV(txid repository.TxID, egID cmap.ID, role int, volID, moveTo cmap.ID) error
	VolEGIncr(txid repository.TxID, vID cmap.ID) error
	VolEGDecr(txid repository.TxID, vID cmap.ID) error
	FindAllChunks(egID cmap.ID, status string) ([]int, error)
}

func newJobRepository(r Repository) jobRepository {
	return r
}
