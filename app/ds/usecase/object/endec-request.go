package object

import (
	"sync"

	"github.com/chanyoung/nil/pkg/client"
)

type request struct {
	r   client.RequestEvent
	wg  sync.WaitGroup
	err error
}

func newRequest(r client.RequestEvent) *request {
	return &request{
		r: r,
	}
}

func (r *request) wait() error {
	r.wg.Wait()

	return r.err
}
