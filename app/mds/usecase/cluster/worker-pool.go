package cluster

import (
	"fmt"
	"sync"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
)

// WorkerPool is a service for managing worker. It is involved in two entities,
// a job entity and a worker entity. It fetches a job from the job repository
// and dispatches the worker.
type workerPool struct {
	// Number of available workers.
	pool map[ID]*worker

	// urgent is the channel to scheduler.
	// Send the urgent job to dispatch it first.
	urgent chan *Job

	cmapAPI cmap.MasterAPI
	store   jobRepository
	mu      sync.Mutex
}

// newWorkerPool returns a new worker pool service.
func newWorkerPool(numWorker int, cmapAPI cmap.MasterAPI, store jobRepository) *workerPool {
	p := &workerPool{
		pool:    make(map[ID]*worker, numWorker),
		urgent:  make(chan *Job, 3),
		cmapAPI: cmapAPI,
		store:   store,
	}

	for i := 0; i < numWorker; i++ {
		w := newWorker(ID(i), cmapAPI, store)
		p.pool[w.id] = w
	}
	go p.runScheduler()

	return p
}

// runScheduler starts to run the worker pool service scheduler.
// Pool scheduler fetches job from the repository and dispatch it repeatedly.
func (p *workerPool) runScheduler() {
	// Check the job repository at least every 10 second.
	checkTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-checkTicker.C:
			p.schedule(nil)
		case job := <-p.urgent:
			p.schedule(job)
		}
	}
}

// dispatchNow send the notification to workPool scheduler with the urgent job.
// (which is usually interactive job)
func (p *workerPool) dispatchNow(j *Job) {
	p.urgent <- j
}

// schedule fetch a job from the job repository and assign it to worker.
func (p *workerPool) schedule(job *Job) {
	txid, err := p.store.Begin()
	if err != nil {
		return
	}

	if job == nil {
		job, err = p.fetchJob(txid)
		if err != nil {
			p.store.Rollback(txid)
			return
		}
	}

	w, ok := p.fetchWorker(job.Type == Batch)
	if ok == false {
		p.store.Rollback(txid)
		return
	}

	if err = p.store.Commit(txid); err != nil {
		p.store.Rollback(txid)
		return
	}

	go w.run(job)
}

func (p *workerPool) fetchJob(txid repository.TxID) (*Job, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *workerPool) fetchWorker(isBatch bool) (fetched *worker, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, w := range p.pool {
		if w.state != idle {
			continue
		}

		w.state = working
		return w, true
	}

	if isBatch {
		return nil, false
	}

	return newContractWorker(p.cmapAPI, p.store), true
}
