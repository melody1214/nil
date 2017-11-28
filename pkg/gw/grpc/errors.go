package grpc

import (
	"errors"
)

var (
	// ErrNilMux occurs when try to register gRPC handler on the nil mux object.
	ErrNilMux = errors.New("gRPC: cannot register gRPC handler on the nil mux")
)
