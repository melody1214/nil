package client

import "net/http"

// RequestType represents the type of request.
type RequestType string

const (
	// WriteToPrimary means the request is head to primary ds and will write a single object.
	WriteToPrimary RequestType = "WriteToPrimary"
	// WriteToFollower means the request is head to followers and will wirte a single object.
	WriteToFollower RequestType = "WriteToFollower"
	// UnknownType means the type of the request is unknown.
	UnknownType RequestType = "unknown"
)

func (t RequestType) String() string {
	return string(t)
}

// Headers holds the information which needs to handle requests.
type Headers map[string]string

// NewHeaders returns a new client headers.
func NewHeaders() Headers {
	return make(map[string]string)
}

// SetLocalChainID set the id of the local chain.
func (h Headers) SetLocalChainID(id string) {
	h["Local-Chain-Id"] = id
}

// SetVolumeID set the id of the volume.
func (h Headers) SetVolumeID(id string) {
	h["Volume-Id"] = id
}

// SetChunkID set the id of the chunk which will be writed.
func (h Headers) SetChunkID(id string) {
	h["Chunk-Id"] = id
}

// SetMD5 set the value of the md5sum.
func (h Headers) SetMD5(md5sum string) {
	h["Md5"] = md5sum
}

// Request is the client request for rest API calling.
type Request interface {
	Send() (*http.Response, error)
}
