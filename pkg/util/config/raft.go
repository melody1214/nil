package config

// Raft holds info required to set a raft server.
type Raft struct {
	// ElectionTimeout : Follower didn't receives a heartbeat message
	// over a 'election timeout' period, then it starts new election term.
	ElectionTimeout string
}
