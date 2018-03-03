package store

import (
	"sync"
)

// queue is a FIFO queue for handling io requests(struct Call).
type queue struct {
	requests []*Call
	mu       sync.Mutex
}

func newRequestsQueue() *queue {
	return &queue{
		requests: make([]*Call, 0),
	}
}

func (q *queue) push(c *Call) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.requests = append(q.requests, c)
}

func (q *queue) pop() (c *Call) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.requests) == 0 {
		return nil
	}

	c = q.requests[0]
	q.requests = q.requests[1:]

	return
}
