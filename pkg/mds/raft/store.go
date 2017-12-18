package raft

import (
	"sync"

	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/hashicorp/raft"
)

// Store is a mysql store, which stores nil meta data.
// Meta data seperates two types:
// 1. Global meta data is the cluster information and all changes are
// made via Raft consensus,
// 2. Local meta data is the location information of each objects and
// managed only in local cluster by mysql replication.
type Store struct {
	// Configuration.
	raftCfg *config.Raft
	secuCfg *config.Security

	// Raft consensus mechanism.
	raft *raft.Raft

	// Protect the fields in the Server struct.
	mu sync.RWMutex
}

// New creates a Store object.
func New(raftCfg *config.Raft, secuCfg *config.Security) *Store {
	return &Store{
		raftCfg: raftCfg,
		secuCfg: secuCfg,
	}
}
