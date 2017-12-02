package raft

// State is the basic raft state in the paper.
type State uint8

const (
	// Follower responds to RPCs from candidates and leaders.
	Follower State = iota
	// Candidate requests vote to the cluster.
	Candidate
	// Leader of the cluster and the log entries will only flow
	// from the leader to others.
	Leader
)

func (rs State) string() string {
	switch rs {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}
