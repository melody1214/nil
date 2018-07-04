package partstore

import (
	"sync"

	"github.com/chanyoung/nil/app/ds/infrastructure/repository"
)

// queue is a FIFO queue for handling io requests(struct request.Call).
type queue struct {
	requests []*repository.Request
	mu       sync.Mutex
}

func newRequestsQueue() *queue {
	return &queue{
		requests: make([]*repository.Request, 0),
	}
}

func (q *queue) push(r *repository.Request) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.requests = append(q.requests, r)
}

func (q *queue) pop() (r *repository.Request) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.requests) == 0 {
		return nil
	}

	r = q.requests[0]
	q.requests = q.requests[1:]

	return
}
