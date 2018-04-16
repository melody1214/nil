package s3

import (
	"net/http"

	"github.com/chanyoung/nil/pkg/client"
)

type s3request struct {
	request   http.Request
	transport http.Transport
	client    http.Client
}

func (r *s3request) Send() (*http.Response, error) {
	return r.client.Do(&r.request)
}

// NewS3Request creates a new s3 request.
func NewS3Request() (client.Request, error) {
	return &s3request{}, nil
}
