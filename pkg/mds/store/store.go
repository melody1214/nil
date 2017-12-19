package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	// Custom transport layer that can encrypts RPCs.
	transport raft.StreamLayer

	// Protect the fields in the Server struct.
	mu sync.RWMutex
}

// New creates a Store object.
func New(raftCfg *config.Raft, secuCfg *config.Security, transport raft.StreamLayer) *Store {
	return &Store{
		raftCfg:   raftCfg,
		secuCfg:   secuCfg,
		transport: transport,
	}
}

// Open opens the store.
func (s *Store) Open() error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.raftCfg.LocalClusterRegion)

	// Create Raft log store directory.
	if s.raftCfg.RaftDir == "" {
		return errors.New("open raft: no raft storage directory specified")
	}
	os.MkdirAll(s.raftCfg.RaftDir, 0700)

	// Setup Raft communication
	transport := raft.NewNetworkTransport(s.transport, maxPool, timeout, mlog.GetLogger().Writer())

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

	// If LocalClusterAddr is same with GlobalClusterAddr then this node
	// becomes the first node, and therefore leader of the cluster.
	if s.raftCfg.LocalClusterAddr == s.raftCfg.GlobalClusterAddr {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				raft.Server{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	} else {
		// Join to the existing raft cluster.
		if err := s.join(s.raftCfg.GlobalClusterAddr,
			s.raftCfg.LocalClusterAddr,
			s.raftCfg.LocalClusterRegion); err != nil {
			return errors.Wrap(err, "open raft: failed to join existing cluster")
		}
	}

	return nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *Store) Join(nodeID, addr string) error {
	mlog.GetLogger().Infof("received join request for remote node %s at %s", nodeID, addr)

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	mlog.GetLogger().Infof("node %s at %s joined successfully", nodeID, addr)
	return nil
}

// join joins into the existing cluster, located at joinAddr.
// The joinAddr node must the leader state node.
func (s *Store) join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("https://%s/join", joinAddr),
		"application/raft",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
