package raft

import (
	"context"
	"io"

	"github.com/chanyoung/nil/pkg/raft/raftpb"
	"github.com/pkg/errors"
)

// Join resend the raft cluster join request from some follower node to the cluster member.
func (h *raftHandlers) Join(in *raftpb.JoinRequest, stream raftpb.Raft_JoinServer) error {
	cc, err := h.mdsMap.Dial()
	if err != nil {
		return errors.Wrap(err, "join raft cluster failed")
	}

	cli := raftpb.NewRaftClient(cc)

	s, err := cli.Join(context.Background(), &raftpb.JoinRequest{})
	if err != nil {
		return errors.Wrap(err, "join raft cluster failed")
	}

	for {
		r, err := s.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "join raft cluster failed")
		}
		stream.Send(r)
	}

	return nil
}
