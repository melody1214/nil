package raft

import (
	"errors"
)

var (
	// ErrNilMux occurs when try to register raft handler on the nil mux object.
	ErrNilMux = errors.New("raft: cannot register gRPC handler on the nil mux")
)
