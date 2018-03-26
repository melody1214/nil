package encoder

import (
	"strings"

	"github.com/chanyoung/nil/pkg/ds/store"
	"github.com/chanyoung/nil/pkg/ds/store/request"
)

type Encoder struct {
	s      store.Service
	q      *queue
	pushCh chan interface{}
}

func NewEncoder(s store.Service) *Encoder {
	return &Encoder{
		s:      s,
		q:      newRequestsQueue(),
		pushCh: make(chan interface{}, 1),
	}
}

func (e *Encoder) Run() {
	for {
		select {
		case <-e.pushCh:
			e.doAll()
		}
	}
}

func (e *Encoder) Push(r *Request) {
	e.q.push(r)
	r.wg.Add(1)
	e.pushCh <- nil
}

func (e *Encoder) doAll() {
	for {
		if r := e.q.pop(); r != nil {
			e.do(r)
			continue
		}

		break
	}
}

func (e *Encoder) do(r *Request) {
	defer r.wg.Done()

	req := &request.Request{
		Op:  request.Write,
		Vol: r.R.Header.Get("Volume-Id"),
		Oid: strings.Replace(strings.Trim(r.R.RequestURI, "/"), "/", ".", -1),

		In: r.R.Body,
	}

	r.err = e.s.Push(req)
	if r.err != nil {
		return
	}

	r.err = req.Wait()
}
