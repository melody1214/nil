package request

import (
	"fmt"
	"io"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/client/s3"
)

// NewRequest returns a new request.
func NewRequest(reqType client.RequestType, method, url string, body io.Reader, headers client.Headers, contentLength int64, opts ...Option) (client.Request, error) {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	if o.useS3 {
		return s3.NewS3Request(reqType, method, url, body, contentLength, headers)
	}

	return nil, fmt.Errorf("no matching client type")
}
