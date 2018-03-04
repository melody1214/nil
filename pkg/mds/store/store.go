package store

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chanyoung/nil/pkg/mds/mysql"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/pkg/errors"
)

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
	db *mysql.MySQL

	// Custom transport layer that can encrypts RPCs.
	transport raft.StreamLayer

	// Protect the fields in the Server struct.
	mu sync.RWMutex
}

// New creates a Store object.
func New(cfg *config.Mds, transport raft.StreamLayer) *Store {
	return &Store{
		cfg:       cfg,
		transport: transport,
	}
}

// Open opens the store.
func (s *Store) Open() error {
	// Connect and initiate to mysql server.
	db, err := mysql.New(s.cfg)
	if err != nil {
		return err
	}
	s.db = db

	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.cfg.Raft.LocalClusterRegion)
	config.LogOutput = mlog.GetLogger().Writer()

	// Create Raft log store directory.
	if s.cfg.Raft.RaftDir == "" {
		return errors.New("open raft: no raft storage directory specified")
	}
	os.MkdirAll(s.cfg.Raft.RaftDir, 0755)

	// Setup Raft communication
	transport := raft.NewNetworkTransport(s.transport, maxPool, timeout, mlog.GetLogger().Writer())

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.cfg.Raft.RaftDir, retainSnapshotCount, mlog.GetLogger().Writer())
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
		return s.addRegion(
			s.cfg.Raft.LocalClusterRegion,
			s.cfg.Raft.LocalClusterAddr,
		)
	}

	return nil
}

// Close cleans up the store.
func (s *Store) Close() {
	// Close mysql connection.
	s.db.Close()
	s.raft.Shutdown()
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *Store) Join(nodeID, addr string) error {
	mlog.GetLogger().Infof("received join request for remote node %s at %s", nodeID, addr)

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

	mlog.GetLogger().Infof("node %s at %s joined successfully", nodeID, addr)
	return nil
}
