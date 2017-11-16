package s3

import (
	"errors"
)

var (
	// ErrNilMux occurs when try to register s3 handler on the nil mux object.
	ErrNilMux = errors.New("s3: cannot register s3 handler on the nil mux")
)
