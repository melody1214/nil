package mysql

import (
	"sync"
	"time"

	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

const (
	retainSnapshotCount = 2
	maxPool             = 3
	timeout             = 10 * time.Second
)

// Store is a mysql store, which stores nil meta data.
// Meta data separates two types:
// 1. Global meta data is the cluster information and all changes are
// made via Raft consensus,
// 2. Local meta data is the location information of each objects and
// managed only in local cluster by mysql replication.
type Store struct {
	// Configuration.
	cfg *config.Mds

	// raft service.
	rs *raftService

	// Mysql store.
	db *mySQL

	// Protect the fields in the store struct.
	mu sync.RWMutex
}

// New creates a Store object.
func New(cfg *config.Mds) *Store {
	logger = mlog.GetPackageLogger("app/mds/infrastructure/repository/mysql")

	s := &Store{cfg: cfg}
	s.rs = newRaftService(s)

	return s
}

// Begin returns a transaction ID.
func (s *Store) Begin() (txid repository.TxID, err error) {
	return s.db.begin()
}

// Rollback rollbacks the transaction with the given ID.
func (s *Store) Rollback(txid repository.TxID) error {
	return s.db.rollback(txid)
}

// Commit commits the transaction with the given ID.
// Auto remove the transaction only when the transaction has been succeeded.
func (s *Store) Commit(txid repository.TxID) error {
	return s.db.commit(txid)
}
