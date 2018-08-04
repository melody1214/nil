package chunk

import "os"

type ObjectHandleBase struct {
	Header Header
	chunk  Name
}

type ObjectHeader struct {
	Name   [48]byte
	MD5    [32]byte
	Size   int64
	Offset int64
}

type ObjectReader interface {
	Read(object Name) (*os.File, error)
}

type ObjectWriter interface {
	Write(object Name) (*os.File, error)
}

type ObjectHandle interface {
	NewReader() ObjectReader
	NewWriter() ObjectWriter
}
