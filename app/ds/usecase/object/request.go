package object

import (
	"net/http"
	"sync"
)

type request struct {
	r   *http.Request
	wg  sync.WaitGroup
	err error
}

func newRequest(r *http.Request) *request {
	return &request{
		r: r,
	}
}

func (r *request) wait() error {
	r.wg.Wait()

	return r.err
}
