package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

type fsm Store

type fsmExecuteResponse struct {
	result sql.Result
	err    error
}

// Apply applies a Raft log entry to the store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Errorf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "execute":
		r, err := f.db.execute(c.Query)
		return &fsmExecuteResponse{result: r, err: err}
	default:
		panic(fmt.Errorf("unrecognized command op: %s", c.Op))
	}
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
