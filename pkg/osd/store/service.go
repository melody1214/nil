package store

import (
	"io"
	"os"
	"time"
)

// Service is the backend store service.
type Service struct {
	basePath     string
	requestQueue queue
}

// NewService returns a new backend store service.
func NewService(path string) *Service {
	return &Service{
		basePath: path,
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
	f, err := os.Open(s.basePath + "/" + c.oid)
	if err != nil {
		c.err = err
		return
	}
	defer f.Close()

	_, c.err = io.Copy(c.out, f)
}

func (s *Service) write(c *Call) {
	f, err := os.Create(s.basePath + "/" + c.oid)
	if err != nil {
		c.err = err
		return
	}
	defer f.Close()

	_, c.err = io.Copy(f, c.in)
}

func (s *Service) delete(c *Call) {
	c.err = os.Remove(s.basePath + "/" + c.oid)
}
