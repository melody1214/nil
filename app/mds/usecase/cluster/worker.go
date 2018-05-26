package cluster

import (
	"sync"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

type workerState int

const (
	idle workerState = iota
	working
)

// worker handles job. The worker processes one task at a time and processing
// like finite state machine. A walker has a unique identity and is an entity.
type worker struct {
	id                ID
	state             workerState
	job               *Job
	localParityShards int
	cmapAPI           cmap.MasterAPI
	store             jobRepository
	mu                sync.Mutex
}

func newWorker(id ID, localParityShards int, cmapAPI cmap.MasterAPI, store jobRepository) *worker {
	return &worker{
		id:                id,
		state:             idle,
		localParityShards: localParityShards,
		cmapAPI:           cmapAPI,
		store:             store,
	}
}

// newContractWorker returns a new contract worker object.
// Contract workers only handle the interactive type of job and will be
// deleted from the pool when the job is done. Contract workers are created
// only when there is no regular workers for handling the interactive job.
func newContractWorker(cmapAPI cmap.MasterAPI, store jobRepository) *worker {
	return &worker{
		state:   working,
		cmapAPI: cmapAPI,
		store:   store,
	}
}

// fsm : finite state machine.
// Cluster domain accepts any fail or membership changed notification from
// other domains. The rpc handler will add the notification into the noti queues.
// Then the fsm, who has the responsibility for handling each notifications will
// take noti from queues and handle it with proper state transitioning.
type fsm func() (next fsm)

// run is the engine of dispatched worker.
// Manage the state transitioning until meet the state nil.
func (w *worker) run(j *Job) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.job = j

	startState := w.init
	for state := startState; state != nil; {
		state = state()
	}

	w.job = nil
	w.state = idle
}

// init is the state for initiating the worker.
// Read the job and determine what action should be taken.
func (w *worker) init() fsm {
	switch w.job.Event.Type {
	case LocalJoin:
		return w.ljStart
	case RegisterVolume:
		return w.rvStart
	case Rebalance:
		return w.rbStart
	case Fail:
		return nil
	case Update:
		return w.upStart
	default:
		return nil
	}
}

/* Worker common methods. */

// updateClusterMap updates the cluster map based on the database information.
func (w *worker) updateClusterMap() error {
	txid, err := w.store.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}

	if err = w._updateClusterMap(txid); err != nil {
		w.store.Rollback(txid)
		return err
	}

	if err = w.store.Commit(txid); err != nil {
		w.store.Rollback(txid)
		return err
	}

	return nil
}

func (w *worker) _updateClusterMap(txid repository.TxID) error {
	// Set new map version.
	ver, err := w.store.GetNewClusterMapVer(txid)
	if err != nil {
		return errors.Wrap(err, "failed to set new cmap map version")
	}

	// Create a cmap map with the new version.
	cm, err := w.createClusterMap(ver, txid)
	if err != nil {
		return errors.Wrap(err, "failed to create cmap map")
	}

	return w.cmapAPI.UpdateCMap(cm)
}

func (w *worker) createClusterMap(ver cmap.Version, txid repository.TxID) (*cmap.CMap, error) {
	nodes, err := w.store.FindAllNodes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map nodes")
	}

	vols, err := w.store.FindAllVolumes(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map volumes")
	}

	encGrps, err := w.store.FindAllEncGrps(txid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cmap map encoding groups")
	}

	return &cmap.CMap{
		Version: ver,
		Time:    cmap.Now(),
		Nodes:   nodes,
		Vols:    vols,
		EncGrps: encGrps,
	}, nil
}
