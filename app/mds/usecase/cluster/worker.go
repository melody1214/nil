package cluster

import (
	"sync"
)

type workerState int

const (
	idle workerState = iota
	working
)

// worker handles job. The worker processes one task at a time and processing
// like finite state machine. A walker has a unique identity and is an entity.
type worker struct {
	id    ID
	state workerState
	job   *Job
	store jobRepository
	mu    sync.Mutex
}

func newWorker(id ID, store jobRepository) *worker {
	return &worker{
		id:    id,
		state: idle,
		store: store,
	}
}

// newContractWorker returns a new contract worker object.
// Contract workers only handle the interactive type of job and will be
// deleted from the pool when the job is done. Contract workers are created
// only when there is no regular workers for handling the interactive job.
func newContractWorker(store jobRepository) *worker {
	return &worker{
		state: working,
		store: store,
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
	w.job = j

	startState := w.init
	for state := startState; state != nil; {
		state = state()
	}
}

// init is the state for initiating the worker.
// Read the job and determine what action should be taken.
func (w *worker) init() fsm {
	switch w.job.Event.Type {
	case LocalJoin:
		return w.ljStart
	case Fail:
		return nil
	default:
		return nil
	}
}
