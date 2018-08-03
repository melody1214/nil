package chunk

import (
	"errors"
	"io"
)

// ErrChunkNotExist is used when the chunk not exists.
var ErrChunkNotExist = errors.New("Chunk not exists")

type HandleBase struct {
	Header Header
}

type Header struct {
	Name   [48]byte
	Bucket [48]byte
	MD5    [32]byte
	Size   int64
	Offset int64
}

type Handle interface {
	NewReader() Reader
	NewWriter() Writer
	Object() ObjectHandle
}

type Reader interface {
	Read(chunk Name, writer *io.PipeWriter) error
}

type Writer interface {
	Write(chunk Name, writer *io.PipeWriter) error
	Truncate(chunk Name) error
}

type Name string

func (n Name) String() string {
	return string(n)
}

type Repository interface {
	Find(chunk Name) (Handle, error)
	Create(chunk Name) (Handle, error)
}
