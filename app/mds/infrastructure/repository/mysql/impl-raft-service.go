package mysql

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	raftdomain "github.com/chanyoung/nil/app/mds/domain/service/raft"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/pkg/errors"
)

var (
	// ErrRaftNotOpened is used when try to access the not opened raft service.
	ErrRaftNotOpened = errors.New("raft service: is not opened")

	// ErrRaftAlreadyOpened is used when try to open the opened raft service.
	ErrRaftAlreadyOpened = errors.New("raft service: is already opened")

	// ErrRaftInvalidArgs is used when the arguments are invalid.
	ErrRaftInvalidArgs = errors.New("raft service: invalid arguments")

	// ErrRaftNotLeader is used when the current node is not the leader of the cluster.
	ErrRaftNotLeader = errors.New("raft service: not leader")

	// ErrRaftInternal is used when the internal error is occured in the raft service.
	ErrRaftInternal = errors.New("raft service: internal error")
)

type raftService struct {
	// Parent database.
	store *Store

	// Raft consensus mechanism.
	raft *raft.Raft

	// Custom transport layer that can encrypts RPCs.
	transport raft.StreamLayer

	// Set true if raft is opened.
	opened bool
	mu     sync.Mutex
}

func newRaftService(s *Store) *raftService {
	return &raftService{
		store:  s,
		opened: false,
	}
}

// NewRaftService returns the raft domain service object.
func (s *Store) NewRaftService() raftdomain.Service {
	return s.rs
}

func (s *raftService) Open(raftL *nilmux.Layer) error {
	ctxLogger := mlog.GetMethodLogger(logger, "raftService.Open")

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.opened {
		ctxLogger.Error(ErrRaftAlreadyOpened)
		return ErrRaftAlreadyOpened
	}

	s.transport = nilmux.NewRaftTransportLayer(raftL)

	// Connect and initiate to mysql server.
	db, err := newMySQL(s.store.cfg)
	if err != nil {
		return err
	}
	s.store.db = db

	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.store.cfg.Raft.LocalClusterRegion)
	config.LogOutput = logger.Writer()
	config.HeartbeatTimeout = 5000 * time.Millisecond
	config.ElectionTimeout = 5000 * time.Millisecond
	config.CommitTimeout = 500 * time.Millisecond
	config.LeaderLeaseTimeout = 5000 * time.Millisecond

	// Create Raft log store directory.
	if s.store.cfg.Raft.RaftDir == "" {
		ctxLogger.Error(errors.Wrap(err, "empty raft directory path"))
		return ErrRaftInvalidArgs
	}
	os.MkdirAll(s.store.cfg.Raft.RaftDir, 0755)

	// Setup Raft communication
	transport := raft.NewNetworkTransport(s.transport, maxPool, timeout, logger.Writer())

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.store.cfg.Raft.RaftDir, retainSnapshotCount, logger.Writer())
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to make new snapshot store"))
		return ErrRaftInternal
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.store.cfg.Raft.RaftDir, "raft.db"))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to make new boltdb store"))
		return ErrRaftInternal
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s.store), logStore, logStore, snapshots, transport)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to make new raft"))
		return ErrRaftInternal
	}
	s.raft = ra
	s.opened = true

	// If LocalClusterAddr is same with GlobalClusterAddr then this node
	// becomes the first node, and therefore leader of the cluster.
	if s.store.cfg.Raft.LocalClusterAddr == s.store.cfg.Raft.GlobalClusterAddr {
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
		if err := s.store.addRegion(
			s.store.cfg.Raft.LocalClusterRegion,
			s.store.cfg.Raft.LocalClusterAddr,
		); err != nil {
			return err
		}

		return s.store.setGlobalClusterConf()
	}

	return nil
}

func (s *raftService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.opened {
		return ErrRaftNotOpened
	}

	s.store.db.close()
	s.raft.Shutdown()

	s.opened = false
	return nil
}

func (s *raftService) Join(nodeID, addr string) error {
	ctxLogger := mlog.GetMethodLogger(logger, "raftService.Join")
	ctxLogger.Infof("received join request for remote node %s at %s", nodeID, addr)

	if s.raft.State() != raft.Leader {
		return ErrRaftNotLeader
	}
	if !s.opened {
		return ErrRaftNotOpened
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	if err := s.store.addRegion(nodeID, addr); err != nil {
		return err
	}

	ctxLogger.Infof("node %s at %s joined successfully", nodeID, addr)

	return nil
}
