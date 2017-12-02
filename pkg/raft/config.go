package raft

import (
	"time"

	"github.com/chanyoung/nil/pkg/util/config"
)

var (
	// Follower didn't receives a heartbeat message over a 'election timeout'
	// period, then it starts new election term.
	electionTimeout time.Duration
)

func init() {
	et, e := time.ParseDuration(config.Get("raft.election_timeout"))
	if e != nil {
		panic(e)
	}

	electionTimeout = et
}
