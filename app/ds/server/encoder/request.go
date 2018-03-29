package encoder

import (
	"net/http"
	"sync"
)

type Request struct {
	R   *http.Request
	wg  sync.WaitGroup
	err error
}

func NewRequest(r *http.Request) *Request {
	return &Request{
		R: r,
	}
}

func (r *Request) Wait() error {
	r.wg.Wait()

	return r.err
}
