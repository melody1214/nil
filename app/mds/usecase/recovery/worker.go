package recovery

import (
	"fmt"
	"sync/atomic"

	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

// recoveryWorker works for recovering all kinds of failure in the cluster.
// worker struct has all required references and channels for handling its job.
type recoveryWorker struct {
	cfg           *config.Mds
	store         Repository
	cmapAPI       cmap.SlaveAPI
	recoveryJobs  map[jobID]recoveryJob
	rebalanceJobs map[jobID]rebalanceJob

	recoveryCh  chan interface{}
	rebalanceCh chan interface{}
	stopCh      chan interface{}
	stopped     uint32
}

// newWorker returns a new worker.
// Note: recovery domain must have only one working worker.
func newWorker(cfg *config.Mds, cmapAPI cmap.SlaveAPI, store Repository) (*recoveryWorker, error) {
	if cfg == nil || cmapAPI == nil || store == nil {
		return nil, fmt.Errorf("invalid arguments")
	}

	return &recoveryWorker{
		cfg:           cfg,
		cmapAPI:       cmapAPI,
		store:         store,
		recoveryJobs:  make(map[jobID]recoveryJob),
		rebalanceJobs: make(map[jobID]rebalanceJob),

		recoveryCh:  make(chan interface{}, 1),
		rebalanceCh: make(chan interface{}, 1),
		stopCh:      make(chan interface{}, 1),
		stopped:     uint32(1),
	}, nil
}

// For state transition.
// Recovery domain accepts any fail or membership changed notification from
// other domains. The rpc handler will add the notification into the noti queues.
// Then the fsm, who has the responsibility for handling each notifications will
// take noti from queues and handle it with proper state transitioning.
type fsm func() (next fsm)

// run is the engine of dispatched worker.
// Manage the state transitioning until meet the state nil.
func (w *recoveryWorker) run() {
	startState := w.init
	for state := startState; state != nil; {
		state = state()
	}
}

// init is the state for initiating the recovery worker.
// Checks if another worker is already running.
func (w *recoveryWorker) init() fsm {
	// Do not go to the workerStop state.
	// workerStop is a cleanup state for worked worker.
	if w.canRun() == false {
		return nil
	}

	return w.listen
}

// canRun swap stopped variable to running state(1) atomically.
// If worker is already running, then return false.
func (w *recoveryWorker) canRun() bool {
	return atomic.SwapUint32(&w.stopped, uint32(0)) == 1
}

// listen is the state for listening notifications from the outside.
func (w *recoveryWorker) listen() fsm {
	select {
	case <-w.recoveryCh:
		return w.recover
	case <-w.rebalanceCh:
		return w.rebalance
	case <-w.stopCh:
		return w.stop
	}
}

// recover is the state for recovering the failure.
func (w *recoveryWorker) recover() fsm {
	// ctxLogger := mlog.GetMethodLogger(logger, "worker.recover")

	// // Updates membership.
	// w.updateMembership()

	// // Get the new version of cluster map.
	// if err := w.updateClusterMap(); err != nil {
	// 	ctxLogger.Error(errors.Wrap(err, "failed to update cluster map"))
	// }

	return w.listen
}

func (w *recoveryWorker) checkRunningRecoveryJobs() fsm {
	if len(w.recoveryJobs) == 0 {
		return w.makeRecoveryJobs
	}
	return w.fixAffectedRecoveryJobs
}

func (w *recoveryWorker) fixAffectedRecoveryJobs() fsm {
	// TODO: implement.
	return w.makeRecoveryJobs
}

func (w *recoveryWorker) makeRecoveryJobs() fsm {
	// TODO: implement.
	return w.checkPendingRecoveryCh
}

func (w *recoveryWorker) checkPendingRecoveryCh() fsm {
	select {
	case <-w.recoveryCh:
		return w.checkRunningRecoveryJobs
	case <-time.After(0):
		return w.dispatchRecoveryJobs
	}
}

func (w *recoveryWorker) dispatchRecoveryJobs() fsm {
	// TODO: implement.
	return w.listen
}

// rebalance is the state for rebalancing the cluster.
func (w *recoveryWorker) rebalance() fsm {
	ctxLogger := mlog.GetMethodLogger(logger, "worker.rebalance")

	vols, err := w.store.FindAllVolumes(repository.NotTx)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to find all volumes"))
		return w.listen
	}

	if w.needRebalance(vols) == false {
		ctxLogger.Info("no need rebalance")
		return w.listen
	}

	ctxLogger.Info("do rebalance")
	if err := w.rebalanceWithinSameVolumeSpeedGroup(vols); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to rebalance same volume speed group"))
		return w.listen
	}

	if err := w.updateClusterMap(); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to update cmap"))
		return w.listen
	}

	return w.listen
}

func (w *recoveryWorker) checkRunningRebalanceJobs() fsm {
	if len(w.rebalanceJobs) == 0 {
		return w.makeRebalanceJobs
	}
	return w.cancelAffectedRebalanceJobs
}

func (w *recoveryWorker) cancelAffectedRebalanceJobs() fsm {
	// TODO: implement.
	return w.makeRebalanceJobs
}

func (w *recoveryWorker) makeRebalanceJobs() fsm {
	// TODO: implement.
	return w.checkPendingRebalanceCh
}

func (w *recoveryWorker) checkPendingRebalanceCh() fsm {
	select {
	case <-w.rebalanceCh:
		return w.checkRunningRebalanceJobs
	case <-time.After(0):
		return w.dispatchRebalanceJobs
	}
}

func (w *recoveryWorker) dispatchRebalanceJobs() fsm {
	// TODO: implement.
	return w.listen
}

// stop is the state for cleanup and stop the worker.
func (w *recoveryWorker) stop() fsm {
	atomic.SwapUint32(&w.stopped, uint32(1))
	return nil
}
