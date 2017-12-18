package store

import (
	"io"

	"github.com/hashicorp/raft"
)

type fsm Store

func (f *fsm) Apply(l *raft.Log) interface{} {
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	return nil
}

type fsmSnapshot struct {
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (f *fsmSnapshot) Release() {}
