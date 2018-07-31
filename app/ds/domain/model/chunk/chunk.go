package chunk

import "io"

type Prefix string

const (
	Writing   Prefix = "W"
	Truncated Prefix = "T"
)

type ChunkHandle struct {
	Header        Header
	volume        Name
	encodingGroup Name
}

func (h *ChunkHandle) Volume() Name {
	return h.volume
}

func (h *ChunkHandle) EncodingGroup() Name {
	return h.encodingGroup
}

type Header struct {
	Name   [48]byte
	Bucket [48]byte
	MD5    [32]byte
	Size   int64
	Offset int64
}

func (h ChunkHandle) NameToStr() string {
	return string(h.Header.Name[:])
}

func (h ChunkHandle) BucketToStr() string {
	return string(h.Header.Bucket[:])
}

func (h ChunkHandle) MD5ToStr() string {
	return string(h.Header.MD5[:])
}

type Name string

func (n Name) String() string {
	return string(n)
}

func New(chunk, bucket, volume, encodingGroup Name) *ChunkHandle {
	h := &ChunkHandle{
		Header: Header{
			Size:   0,
			Offset: 0,
		},
		volume:        volume,
		encodingGroup: encodingGroup,
	}
	copy(h.Header.Name[:], chunk.String())
	copy(h.Header.Bucket[:], bucket.String())

	return h
}

type Repository interface {
	NewReader(*ChunkHandle) io.Reader
	NewWriter(*ChunkHandle) io.Writer
	Find(chunk Name) (*ChunkHandle, error)
}
