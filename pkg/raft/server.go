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
		fmt.Printf("Message: %#v\n", r)
	}

	return nil
}
