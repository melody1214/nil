package cluster

import (
	"fmt"
	"sync"
)

// WorkerPool is a service for managing worker. It is involved in two entities,
// a job entity and a worker entity. It fetches a job from the job repository
// and dispatches the worker.
type workerPool struct {
	// Number of available workers.
	available int

	pool map[ID]*worker

	// tempPool is a worker pool of contract workers.
	// Contract workers only handle the interactive type of job and will be
	// deleted from the pool when the job is done.
	// Contract workers are created only when there is no regular workers
	// for handling the interactive job.
	tempPool map[ID]*worker

	mu sync.Mutex
}

// newWorkerPool returns a new worker pool service.
func newWorkerPool(numWorker int) (*workerPool, error) {
	if numWorker <= 0 {
		return nil, fmt.Errorf("invalid number of workers")
	}
	p := make(map[ID]*worker, numWorker)

	for i := 0; i < numWorker; i++ {
		w := newWorker(ID(i))
		p[w.id] = w
	}

	return &workerPool{
		available: numWorker,
		pool:      p,
	}, nil
}

// run starts to run the worker pool service.
// Pool service fetches job from the repository and dispatch it repeatedly.
func (p *workerPool) run() {
}
