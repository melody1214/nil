package store

import (
	"fmt"
	"io"
	"sync"

	context "golang.org/x/net/context"
)

// Operation indicates what backend store operation is called.
type Operation int

const (
	// Read requests an object handle that can read the requested object.
	Read Operation = iota
	// Write requests an object handle that can write the requested object.
	Write
	// Delete requests an object handle that can delete the requested object.
	Delete
)

// Call includes information about backend store request.
type Call struct {
	s   *Service
	op  Operation
	lv  string
	oid string

	in  io.Reader
	out io.Writer

	wg  sync.WaitGroup
	ctx context.Context

	// Call result, set after method 'DO' called.
	err error
}

// Call returns a Call structure, which is used for backend store IO request.
func (s *Service) Call() *Call {
	return &Call{
		s: s,
	}
}

// Operation sets an io operation type of the call.
func (c *Call) Operation(op Operation) *Call {
	c.op = op
	return c
}

// Lv sets the logical volume name that stores the target object.
func (c *Call) Lv(lv string) *Call {
	c.lv = lv
	return c
}

// ObjectID sets the target object ID.
func (c *Call) ObjectID(objectID string) *Call {
	c.oid = objectID
	return c
}

// InputStream sets the data input stream for writing call.
func (c *Call) InputStream(in io.Reader) *Call {
	c.in = in
	return c
}

// OutputStream sets the data output stream for reading call.
func (c *Call) OutputStream(out io.Writer) *Call {
	c.out = out
	return c
}

// Context sets the context to be used in this call's Do method.
func (c *Call) Context(ctx context.Context) *Call {
	c.ctx = ctx
	return c
}

// Do executes the backend store io request call.
func (c *Call) Do() error {
	if c.s == nil {
		return fmt.Errorf("%v: invalid arguments", c)
	}

	switch c.op {
	case Read:
		if c.lv == "" || c.oid == "" || c.out == nil {
			return fmt.Errorf("%v: invalid arguments", c)
		}
	case Write:
		if c.lv == "" || c.oid == "" || c.in == nil {
			return fmt.Errorf("%v: invalid arguments", c)
		}
	case Delete:
		if c.lv == "" || c.oid == "" {
			return fmt.Errorf("%v: invalid arguments", c)
		}
	default:
		return fmt.Errorf("%v: invalid arguments", c)
	}

	// TODO: change to use context, not wg.
	c.wg.Add(1)

	c.s.requestQueue.push(c)

	c.wg.Wait()

	return c.err
}
