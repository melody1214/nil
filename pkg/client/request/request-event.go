package request

import "net/http"
import "github.com/chanyoung/nil/pkg/client"
import "github.com/chanyoung/nil/pkg/client/s3"

// NewRequestEventFactory returns a new request event factory.
func NewRequestEventFactory(opts ...EventFactoryOption) *RequestEventFactory {
	f := &RequestEventFactory{
		o: defaultEventFactoryOptions,
	}

	for _, opt := range opts {
		opt(&f.o)
	}

	return f
}

// RequestEventFactory creates handles for request event.
type RequestEventFactory struct {
	o eventFactoryOptions
}

// CreateRequestEvent creates a validated request event.
func (f *RequestEventFactory) CreateRequestEvent(w http.ResponseWriter, r *http.Request) (client.RequestEvent, error) {
	switch classifyProtocol(r.Header) {
	case client.S3:
		return s3.NewS3RequestEvent(w, r)
	default:
		return nil, client.ErrInvalidProtocol
	}
}

func classifyProtocol(h http.Header) client.Protocol {
	if ok := h.Get("X-Amz-Date"); ok != "" {
		return client.S3
	}
	if ok := h.Get("Amz-Sdk-Invocation-Id"); ok != "" {
		return client.S3
	}

	return client.Unknown
}
