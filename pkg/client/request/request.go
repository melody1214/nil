package request

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/client/s3"
)

// NewRequest returns a new request.
func NewRequest(opts ...Option) (client.Request, error) {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	if o.useS3 {
		return s3.NewS3Request()
	}

	return nil, fmt.Errorf("no matching client type")
}
