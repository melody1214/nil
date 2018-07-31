package object

import (
	"io"
)

type ObjectHandle struct {
	Header Header
	chunk  Name
}

type Header struct {
	Name   [48]byte
	MD5    [32]byte
	Size   int64
	Offset int64
}

func (h ObjectHandle) NameToStr() string {
	return string(h.Header.Name[:])
}

func (h ObjectHandle) MD5ToStr() string {
	return string(h.Header.MD5[:])
}

type Name string

func (n Name) String() string {
	return string(n)
}

func New(object, volume, encodingGroup, chunk Name) *ObjectHandle {
	h := &ObjectHandle{
		Header: Header{
			Size:   0,
			Offset: 0,
		},
		chunk: chunk,
	}
	copy(h.Header.Name[:], chunk.String())

	return h
}

type Repository interface {
	NewReader(*ObjectHandle) io.Reader
	NewWriter(*ObjectHandle) io.Writer
	Find(object, chunk Name) (*ObjectHandle, error)
}
