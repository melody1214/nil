package cluster

import "sync"

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
	job   Job
	mu    sync.Mutex
}

func newWorker(id ID) *worker {
	return &worker{
		id:    id,
		state: idle,
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
func (w *worker) run(j Job) {
	if w.canStart() == false {
		return
	}

	startState := w.init
	for state := startState; state != nil; {
		state = state()
	}
}

func (w *worker) canStart() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state != idle {
		return false
	}

	w.state = working
	return true
}

// init is the state for initiating the worker.
// Read the job and determine what action should be taken.
func (w *worker) init() fsm {
	switch w.job.Event.Type {
	case AddNode:
		return nil
	case Fail:
		return nil
	default:
		return nil
	}
}
