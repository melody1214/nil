package request

import (
	"fmt"
	"io"
	"net/http"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/client/s3"
	"github.com/pkg/errors"
)

// NewRequest returns a new request.
func NewRequest(reqType client.RequestType, method, url string, body io.Reader, headers client.Headers, contentLength int64, opts ...Option) (client.Request, error) {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	// Create a http request.
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request with the given arguments")
	}
	if contentLength != 0 {
		request.ContentLength = contentLength
	}

	// Set request type into the headers.
	request.Header.Set("Request-Type", reqType.String())

	// Set headers.
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	if o.useS3 {
		return s3.NewS3Request(request, o.genSign, o.cred)
	}

	return nil, fmt.Errorf("no matching client type")
}
