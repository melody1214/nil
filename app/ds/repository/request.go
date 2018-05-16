package repository

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
	// WriteAll requests an object handle that can write the requested object.
	WriteAll
	// Delete requests an object handle that can delete metadata of the requested object.
	Delete
	// ReadAll requests an object handle that can read the requested chunk.
	ReadAll
	// DeleteReal requests an object handle that can delete the requested object.
	DeleteReal
)

// Request includes information abOut backend store request.
type Request struct {
	Op     Operation
	User   string // User ID
	Vol    string // Volume
	LocGid string // Local group ID
	Oid    string // Object ID
	Cid    string // Chunk ID

	Osize int64  // Object Size
	Md5   string // MD5 string

	In  io.Reader
	Out io.Writer

	Wg  sync.WaitGroup
	Ctx context.Context

	// Call result, set after method 'DO' called.
	Err error
}

// Verify verifies the each fields of request are valid.
func (r *Request) Verify() error {
	switch r.Op {
	case Read:
		if r.Vol == "" || r.Oid == "" || r.Out == nil {
			return fmt.Errorf("%v: invalid arguments", r)
		}
	case Write:
		if r.Vol == "" || r.Oid == "" || r.In == nil {
			return fmt.Errorf("%v: invalid arguments", r)
		}
	case Delete:
		if r.Vol == "" || r.Oid == "" {
			return fmt.Errorf("%v: invalid arguments", r)
		}
	case ReadAll:
		if r.Vol == "" || r.Cid == "" || r.Out == nil {
			return fmt.Errorf("%v: invalid arguments", r)
		}
	case WriteAll:
		if r.Vol == "" || r.Cid == "" || r.In == nil {
			return fmt.Errorf("%v: invalid arguments", r)
		}
	case DeleteReal:
		if r.Vol == "" || (r.Oid == "" && r.Cid == "") {
			return fmt.Errorf("%v: invalid arguments", r)
		}
	default:
		return fmt.Errorf("%v: invalid arguments", r)
	}

	return nil
}

// Wait waits until the request is finished.
// TODO: with timeout option.
func (r *Request) Wait() error {
	r.Wg.Wait()

	return r.Err
}
