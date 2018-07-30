package raft

import "github.com/chanyoung/nil/pkg/nilmux"

// Service provides methods for dealing with raft cluster.
type Service interface {
	Open(raftL *nilmux.Layer) error
	Close() error
	Join(nodeID, addr string) error
	NewRaftSimpleService() SimpleService
}

// SimpleService provides basic methods for dealing with raft cluster.
type SimpleService interface {
	Leader() (bool, error)
	LeaderEndPoint() (string, error)
}
