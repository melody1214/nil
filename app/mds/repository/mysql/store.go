package mysql

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/pkg/errors"
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

	// Raft consensus mechanism.
	raft *raft.Raft

	// Mysql store.
	db *mySQL

	// Custom transport layer that can encrypts RPCs.
	transport raft.StreamLayer

	// Protect the fields in the store struct.
	mu sync.RWMutex
}

// New creates a Store object.
func New(cfg *config.Mds) *Store {
	logger = mlog.GetPackageLogger("app/mds/repository/mysql")

	return &Store{
		cfg: cfg,
	}
}

// Open opens the store.
func (s *Store) Open(raftL *nilmux.Layer) error {
	s.transport = nilmux.NewRaftTransportLayer(raftL)

	// Connect and initiate to mysql server.
	db, err := newMySQL(s.cfg)
	if err != nil {
		return err
	}
	s.db = db

	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.cfg.Raft.LocalClusterRegion)
	config.LogOutput = logger.Writer()
	config.HeartbeatTimeout = 5000 * time.Millisecond
	config.ElectionTimeout = 5000 * time.Millisecond
	config.CommitTimeout = 500 * time.Millisecond
	config.LeaderLeaseTimeout = 5000 * time.Millisecond

	// Create Raft log store directory.
	if s.cfg.Raft.RaftDir == "" {
		return errors.New("open raft: no raft storage directory specified")
	}
	os.MkdirAll(s.cfg.Raft.RaftDir, 0755)

	// Setup Raft communication
	transport := raft.NewNetworkTransport(s.transport, maxPool, timeout, logger.Writer())

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.cfg.Raft.RaftDir, retainSnapshotCount, logger.Writer())
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new snapshot store")
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.cfg.Raft.RaftDir, "raft.db"))
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new boltdb store")
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, logStore, snapshots, transport)
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new raft")
	}
	s.raft = ra

	// If LocalClusterAddr is same with GlobalClusterAddr then this node
	// becomes the first node, and therefore leader of the cluster.
	if s.cfg.Raft.LocalClusterAddr == s.cfg.Raft.GlobalClusterAddr {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				raft.Server{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)

		// Waiting until become raft leader.
		for s.raft.State() != raft.Leader {
			time.Sleep(10 * time.Millisecond)
		}
		// Add my region.
		if err := s.addRegion(
			s.cfg.Raft.LocalClusterRegion,
			s.cfg.Raft.LocalClusterAddr,
		); err != nil {
			return err
		}

		return s.setGlobalClusterConf()
	}

	return nil
}

// Close cleans up the store.
func (s *Store) Close() error {
	// Close mysql connection.
	s.db.close()
	s.raft.Shutdown()

	return nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *Store) Join(nodeID, addr string) error {
	ctxLogger := mlog.GetMethodLogger(logger, "store.Join")
	ctxLogger.Infof("received join request for remote node %s at %s", nodeID, addr)

	if s.raft.State() != raft.Leader {
		return errors.New("not leader")
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	if err := s.addRegion(nodeID, addr); err != nil {
		return err
	}

	ctxLogger.Infof("node %s at %s joined successfully", nodeID, addr)
	return nil
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

// leaderEndPoint returns the endpoint of global cluster region.
func (s *Store) leaderEndPoint() (endpoint string) {
	if s.raft == nil {
		return
	}

	future := s.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return
	}

	servers := future.Configuration().Servers
	// Not joined yet.
	if len(servers) == 1 {
		return
	}

	var leader *raft.Server
	leaderAddress := s.raft.Leader()
	for _, s := range servers {
		if s.Address == leaderAddress {
			leader = &s
			break
		}
	}

	if leader == nil {
		return
	}

	q := fmt.Sprintf(
		`
		SELECT rg_end_point
		FROM region
		WHERE rg_name='%s'
		`, string(leader.ID),
	)

	s.QueryRow(repository.NotTx, q).Scan(&endpoint)
	return
}
