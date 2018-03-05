package lvstore

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chanyoung/nil/pkg/ds/store/request"
	"github.com/chanyoung/nil/pkg/ds/store/volume"
)

// Service is the backend store service.
type Service struct {
	lvs          map[string]*lv
	basePath     string
	requestQueue queue
}

// NewService returns a new backend store service.
func NewService(basePath string) *Service {
	return &Service{
		basePath: basePath,
		lvs:      map[string]*lv{},
	}
}

// Run starts to serve backend store service.
func (s *Service) Run() {
	// TODO: change to do not polling.
	for {
		if c := s.requestQueue.pop(); c != nil {
			s.handleCall(c)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// Push pushes an io request into the scheduling queue.
func (s *Service) Push(r *request.Request) error {
	if err := r.Verify(); err != nil {
		return err
	}

	r.Wg.Add(1)
	s.requestQueue.push(r)

	return nil
}

// AddVolume adds a volume into the lv map.
func (s *Service) AddVolume(v *volume.Vol) error {
	return nil
}

func (s *Service) handleCall(r *request.Request) {
	defer r.Wg.Done()

	switch r.Op {
	case request.Read:
		s.read(r)
	case request.Write:
		s.write(r)
	case request.Delete:
		s.delete(r)
	}
}

func (s *Service) read(r *request.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	f, err := os.Open(lv.MntPoint + "/" + r.Oid)
	if err != nil {
		r.Err = err
		return
	}
	defer f.Close()

	_, r.Err = io.Copy(r.Out, f)
}

func (s *Service) write(r *request.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	f, err := os.Create(lv.MntPoint + "/" + r.Oid)
	if err != nil {
		r.Err = err
		return
	}
	defer f.Close()

	_, r.Err = io.Copy(f, r.In)
}

func (s *Service) delete(r *request.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	r.Err = os.Remove(lv.MntPoint + "/" + r.Oid)
}
