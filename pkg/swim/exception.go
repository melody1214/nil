package swim

import (
	"errors"
	"time"
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
	// ErrChanged is a special error message when it occurs the membership has changed.
	ErrChanged = errors.New("swim: membership has changed")
)

// PingError contains specific error information which occurred in swim server.
type PingError struct {
	Type   MethodName
	DestID ServerID
	Err    string
}

// handleErr handles ping errors to check and disseminate status.
func (s *Server) handleErr(pe PingError, pec chan PingError) {
	switch pe.Type {
	case Ping:
		s.suspect(pe.DestID)

		// Keep suspect state for 1minute.
		time.Sleep(1 * time.Minute)

		go s.pingRequest(pe.DestID, pec)
	case PingRequest:
		s.faulty(pe.DestID)
	}
}
