package swim

import (
	"errors"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

var (
	// ErrRunning occurs when try to start swim server which is already running.
	ErrRunning = errors.New("swim: server is already running")
	// ErrStopped occurs when try to stop swim server which is already stopped.
	ErrStopped = errors.New("swim: server is already stopped")
)

// PingError contains specific error information which occured in swim server.
type PingError struct {
	Type   swimpb.Type
	DestID string
	Err    error
}

func (s *Server) handleErr(pe PingError) {
	return
}
