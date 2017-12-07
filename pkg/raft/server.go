package raft

import (
	"context"
	"fmt"
	"io"

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

	// Log store.
	logStore logStore

	// The latest term server has seen.
	currentTerm uint64

	// The highest log entry index which need to be commited.
	commitIndex uint64

	// The highest log entry index which is applied to state machine.
	lastApplied uint64
}

// New creates a raft server object.
func New(raftCfg *config.Raft, secuCfg *config.Security) *Server {
	return &Server{
		raftCfg:     raftCfg,
		secuCfg:     secuCfg,
		state:       Follower,
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

	cc, err := grpc.Dial(s.raftCfg.ClusterAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return errors.Wrap(err, "join raft cluster failed")
	}

	cli := raftpb.NewRaftClient(cc)

	stream, err := cli.Join(context.Background(), &raftpb.JoinRequest{})
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
	index := s.logStore.lastIndex()

	// Log index starts from 1.
	// 0 means log store has no entries.
	var i uint64
	for i = 1; i <= index; i++ {
		l, err := s.logStore.readLog(i)
		if err != nil {
			return err
		}

		if err := stream.Send(&raftpb.JoinResponse{
			MessageType: raftpb.JoinResponse_LOG_MIGRATION,
			LogEntry: &raftpb.LogEntry{
				Index: l.index,
				Query: l.query,
			},
		}); err != nil {
			return err
		}
	}

	// TODO: Sends a join request to leader with the copied log index.

	return nil
}
