package s3

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/pkg/errors"
)

type s3request struct {
	headers     client.Headers
	request     *http.Request
	requestType client.RequestType
	transport   *http.Transport
	client      *http.Client
}

func (r *s3request) Send() (*http.Response, error) {
	return r.client.Do(r.request)
}

// NewS3Request creates a new s3 request.
func NewS3Request(reqType client.RequestType, method, url string, body io.Reader, contentLength int64, headers client.Headers) (client.Request, error) {
	// Create a http request.
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request with the given arguments")
	}
	if contentLength != 0 {
		request.ContentLength = contentLength
	}

	// Set headers.
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	// Set request type into the headers.
	request.Header.Set("Request-Type", reqType.String())

	// Create http transport.
	transport := &http.Transport{
		Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
		TLSClientConfig:     security.DefaultTLSConfig(),
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &s3request{
		headers:     headers,
		request:     request,
		requestType: reqType,
		transport:   transport,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}, nil
}
