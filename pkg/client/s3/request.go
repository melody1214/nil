package s3

import (
	"net/http"

	"github.com/chanyoung/nil/pkg/client"
	s3lib "github.com/chanyoung/nil/pkg/s3"
)

type S3RequestEvent struct {
	protocol client.Protocol

	httpWriter  http.ResponseWriter
	httpRequest *http.Request

	accessKey s3lib.AccessKey
}

func NewS3RequestEvent(w http.ResponseWriter, r *http.Request) (client.RequestEvent, error) {
	return &S3RequestEvent{
		protocol:    client.S3,
		httpWriter:  w,
		httpRequest: r,
	}, nil
}

func (r *S3RequestEvent) Protocol() client.Protocol {
	return r.protocol
}

func (r *S3RequestEvent) ResponseWriter() http.ResponseWriter {
	return r.httpWriter
}

func (r *S3RequestEvent) Request() *http.Request {
	return r.httpRequest
}
