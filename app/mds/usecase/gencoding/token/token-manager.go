package token

import (
	"time"
)

// Manager manages token.
type Manager struct {
	currentVer int
}

// NewManager returns a new object of token manager.
func NewManager() *Manager {
	return &Manager{currentVer: 0}
}

const tokenTimeout = 1 * time.Minute

// NewToken creates a new token object with the given routing legs.
func (m *Manager) NewToken(routing Leg) *Token {
	m.currentVer = m.currentVer + 1

	return &Token{
		Version: m.currentVer,
		Timeout: time.Now().Add(tokenTimeout),
		Routing: routing,
	}
}
