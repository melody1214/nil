package raft

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/chanyoung/nil/pkg/raft/raftpb"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server is the core object of the raft server.
type Server struct {
	// Configuration.
	raftCfg *config.Raft
	secuCfg *config.Security

	// Current server state.
	state State

	// Member servers
	servers map[string]string

	// Log store.
	logStore logStore

	// The latest term server has seen.
	currentTerm uint64

	// The highest log entry index which need to be commited.
	commitIndex uint64

	// The highest log entry index which is applied to state machine.
	lastApplied uint64

	// Protect the fields in the Server struct.
	mu sync.RWMutex
}

// New creates a raft server object.
func New(raftCfg *config.Raft, secuCfg *config.Security) *Server {
	return &Server{
		raftCfg:     raftCfg,
		secuCfg:     secuCfg,
		state:       Follower,
		servers:     make(map[string]string),
		logStore:    newBasicStore(),
		currentTerm: 0,
		commitIndex: 0,
		lastApplied: 0,
	}
}

// Run starts a raft cluster.
func (s *Server) Run(c chan error) {
	if err := s.joinToCluster(); err != nil {
		c <- err
	}
}

func (s *Server) joinToCluster() error {
	creds, err := credentials.NewClientTLSFromFile(
		s.secuCfg.CertsDir+"/"+s.secuCfg.RootCAPem,
		"localhost", // This need to be fixed.
	)
	if err != nil {
		return errors.Wrap(err, "join raft cluster failed")
	}

	cc, err := grpc.Dial(s.raftCfg.GlobalClusterAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return errors.Wrap(err, "join raft cluster failed")
	}

	cli := raftpb.NewRaftClient(cc)

	stream, err := cli.Join(context.Background(), &raftpb.JoinRequest{
		Region:  s.raftCfg.LocalClusterRegion,
		Address: s.raftCfg.LocalClusterAddr,
	})
	if err != nil {
		return errors.Wrap(err, "join raft cluster failed")
	}

	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "join raft cluster failed")
		}
		fmt.Printf("Message: %#v\n%#v", r, r.LogEntry)
	}

	return nil
}

// Join handles the join requests from the follower node.
func (s *Server) Join(in *raftpb.JoinRequest, stream raftpb.Raft_JoinServer) error {
	// Copy logs.
	if err := s.copyLogs(stream); err != nil {
		errors.Wrap(err, "raft join failed")
	}

	// TODO: Sends a join request to leader with the copied log index.
	if err := s.addServer(in); err != nil {
		errors.Wrap(err, "raft join failed")
	}

	return nil
}

func (s *Server) copyLogs(stream raftpb.Raft_JoinServer) error {
	// Index of the highest log entry.
	hIdx := s.logStore.lastIndex()

	// Log index starts from 1.
	// 0 means log store has no entries.
	var idx uint64
	for idx = 1; idx <= hIdx; idx++ {
		l, err := s.logStore.readLog(idx)
		if err != nil {
			return err
		}

		if err := stream.Send(&raftpb.JoinResponse{
			LogEntry: &raftpb.LogEntry{
				Index: l.index,
				Query: l.query,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) addServer(in *raftpb.JoinRequest) error {
	// Add myself.
	if len(s.servers) < 1 {
		// Write log myself.
		// Apply log.
		// return
	}

	// Only leader can add new server into the cluster.
	if s.state != Leader {
		return errors.New("only leader can add new server")
	}

	// Check is duplicated.
	// Write log.
	// Do raft things.
	// return

	return nil
}
