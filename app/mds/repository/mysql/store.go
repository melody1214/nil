package mysql

import (
	"sync"

	"github.com/chanyoung/nil/app/mds/mysql"
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/admin"
	"github.com/chanyoung/nil/app/mds/usecase/auth"
	"github.com/chanyoung/nil/app/mds/usecase/bucket"
	"github.com/chanyoung/nil/app/mds/usecase/clustermap"
	"github.com/chanyoung/nil/app/mds/usecase/membership"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/hashicorp/raft"
)

// store is a mysql store, which stores nil meta data.
// Meta data separates two types:
// 1. Global meta data is the cluster information and all changes are
// made via Raft consensus,
// 2. Local meta data is the location information of each objects and
// managed only in local cluster by mysql replication.
type store struct {
	// Configuration.
	cfg *config.Mds

	// Raft consensus mechanism.
	raft *raft.Raft

	// Mysql store.
	db *mysql.MySQL

	// Custom transport layer that can encrypts RPCs.
	transport raft.StreamLayer

	// Protect the fields in the store struct.
	mu sync.RWMutex
}

// New creates a Store object.
func New(cfg *config.Mds, transport raft.StreamLayer) repository.Store {
	return &store{
		cfg:       cfg,
		transport: transport,
	}
}

// NewAdminRepository returns a new instance of a mysql admin repository.
func NewAdminRepository(s repository.Store) admin.Repository {
	return s
}

// NewAuthRepository returns a new instance of a mysql auth repository.
func NewAuthRepository(s repository.Store) auth.Repository {
	return s
}

// NewBucketRepository returns a new instance of a mysql bucket repository.
func NewBucketRepository(s repository.Store) bucket.Repository {
	return s
}

// NewClusterMapRepository returns a new instance of a mysql cluster map repository.
func NewClusterMapRepository(s repository.Store) clustermap.Repository {
	return s
}

// NewMembershipRepository returns a new instance of a mysql membership repository.
func NewMembershipRepository(s repository.Store) membership.Repository {
	return s
}

// NewObjectRepository returns a new instance of a mysql object repository.
func NewObjectRepository(s repository.Store) object.Repository {
	return s
}

// NewRecoveryRepository returns a new instance of a mysql recovery repository.
func NewRecoveryRepository(s repository.Store) recovery.Repository {
	return s
}
