package raft

import (
	"github.com/chanyoung/nil/pkg/raft/raftpb"
)

// Join resend the raft cluster join request from some follower node to the cluster member.
func (h *raftHandlers) Join(in *raftpb.JoinRequest, stream raftpb.Raft_JoinServer) error {
	testMessage := []raftpb.JoinResponse{
		{
			MessageType: raftpb.JoinResponse_ACK,
			Query:       "query 1",
		},
		{
			MessageType: raftpb.JoinResponse_DB_MIGRATION,
			Query:       "query 2",
		},
		{
			MessageType: raftpb.JoinResponse_LOG_MIGRATION,
			Query:       "query 3",
		},
	}

	for _, m := range testMessage {
		if err := stream.Send(&m); err != nil {
			return err
		}
	}
	return nil
}
