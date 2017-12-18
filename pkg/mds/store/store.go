package store

import (
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

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

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
func (s *Store) Open(enableSingle bool) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.raftCfg.LocalClusterRegion)

	// Create Raft log store directory.
	if s.raftCfg.RaftDir == "" {
		return errors.New("open raft: no raft storage directory specified")
	}
	os.MkdirAll(s.raftCfg.RaftDir, 0700)

	// Setup Raft communication
	addr, err := net.ResolveTCPAddr("tcp", s.raftCfg.BindAddr)
	if err != nil {
		return errors.Wrap(err, "open raft: failed to resolve tcp address")
	}
	transport, err := raft.NewTCPTransport(s.raftCfg.BindAddr, addr, maxPool, timeout, mlog.GetLogger().Writer())
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new tcp transport")
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.raftCfg.RaftDir, retainSnapshotCount, mlog.GetLogger().Writer())
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new snapshot store")
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(s.raftCfg.RaftDir, "raft.db"))
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new boltdb store")
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, logStore, snapshots, transport)
	if err != nil {
		return errors.Wrap(err, "open raft: failed to make new raft")
	}
	s.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				raft.Server{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}
