package chunk

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
}

type Writer interface {
}

type Name string

func (n Name) String() string {
	return string(n)
}

type Repository interface {
	Find(chunk Name) (Handle, error)
	Create(chunk Name) (Handle, error)
}
