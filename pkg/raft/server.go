package raft

import (
	"github.com/chanyoung/nil/pkg/util/config"
)

// Server is the core object of the raft server.
type Server struct {
	// Configuration.
	cfg *config.Raft

	// Current server state.
	state State

	// The latest term server has seen.
	currentTerm uint64

	// The highest log entry index which need to be commited.
	commitIndex uint64

	// The highest log entry index which is applied to state machine.
	lastApplied uint64
}

// New creates a raft server object.
func New(cfg *config.Raft) *Server {
	return &Server{
		cfg:         cfg,
		state:       Follower,
		currentTerm: 0,
		commitIndex: 0,
		lastApplied: 0,
	}
}
