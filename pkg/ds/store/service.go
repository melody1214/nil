package store

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Service is the backend store service.
type Service struct {
	lvs          map[string]*LV
	basePath     string
	requestQueue queue
}

// NewService returns a new backend store service.
func NewService(basePath string) *Service {
	return &Service{
		basePath: basePath,
		lvs:      map[string]*LV{},
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

func (s *Service) handleCall(c *Call) {
	defer c.wg.Done()

	switch c.op {
	case Read:
		s.read(c)
	case Write:
		s.write(c)
	case Delete:
		s.delete(c)
	}
}

func (s *Service) read(c *Call) {
	lv, ok := s.lvs[c.lv]
	if !ok {
		c.err = fmt.Errorf("no such lv: %s", c.lv)
		return
	}

	f, err := os.Open(lv.MntPoint + "/" + c.oid)
	if err != nil {
		c.err = err
		return
	}
	defer f.Close()

	_, c.err = io.Copy(c.out, f)
}

func (s *Service) write(c *Call) {
	lv, ok := s.lvs[c.lv]
	if !ok {
		c.err = fmt.Errorf("no such lv: %s", c.lv)
		return
	}

	f, err := os.Create(lv.MntPoint + "/" + c.oid)
	if err != nil {
		c.err = err
		return
	}
	defer f.Close()

	_, c.err = io.Copy(f, c.in)
}

func (s *Service) delete(c *Call) {
	lv, ok := s.lvs[c.lv]
	if !ok {
		c.err = fmt.Errorf("no such lv: %s", c.lv)
		return
	}

	c.err = os.Remove(lv.MntPoint + "/" + c.oid)
}
