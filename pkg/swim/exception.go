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
	// ErrNotFound occurs when failed to get item from map.
	ErrNotFound = errors.New("swim: item not found")
	// ErrPingReq occurs when failed to retrieve ack from ping requests.
	ErrPingReq = errors.New("swim: failed to retrieve ack from ping requests")
)

// PingError contains specific error information which occurred in swim server.
type PingError struct {
	Type   swimpb.Type
	DestID string
	Err    error
}

// handleErr handles ping errors to check and disseminate status.
func (s *Server) handleErr(pe PingError, pec chan PingError) {
	switch pe.Type {
	case swimpb.Type_PING:
		s.suspect(pe.DestID)
		go s.pingRequest(pe.DestID, pec)
	case swimpb.Type_PINGREQUEST:
		s.faulty(pe.DestID)
	}
}
