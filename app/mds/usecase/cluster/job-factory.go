package cluster

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

// jobFactory is a factory for making job.
// It takes an event as an argument and creates a job with a unique ID.
// Created jobs are stored in the job repository.
type jobFactory struct {
	store jobRepository
}

// newJobFactory returns a new job factory object.
func newJobFactory(s jobRepository) *jobFactory {
	return &jobFactory{
		store: s,
	}
}

// create creates an event with a given event information.
func (f *jobFactory) create(e *Event, private interface{}) (*Job, error) {
	j := &Job{
		Event: *e,
	}

	switch e.Type {
	case LocalJoin:
		j.Type = Iterative
		j.State = Run

		if private == nil {
			return nil, fmt.Errorf("added node information should be in private field")
		}

		n, ok := private.(cmap.Node)
		if ok == false {
			return nil, fmt.Errorf("wrong private field type")
		}
		j.private = n
		j.waitChannel = make(chan error)
		j.Log = newJobLog("request from " + n.Addr.String() + ", name " + n.Name.String())

	case RegisterVolume:
		j.Type = Iterative
		j.State = Run

		if private == nil {
			return nil, fmt.Errorf("register volume information should be in private field")
		}

		v, ok := private.(*cmap.Volume)
		if ok == false {
			return nil, fmt.Errorf("wrong private field type")
		}
		j.private = v
		j.waitChannel = make(chan error)
		j.Log = newJobLog("DS: " + v.ID.String())

	case Rebalance:
		j.Type = Batch
		j.State = Ready

	case Fail:
		j.Type = Batch
		j.State = Ready

	default:
		return nil, fmt.Errorf("unknown event type")
	}

	txid, err := f.store.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction")
	}

	if err = f.store.InsertJob(txid, j); err != nil {
		f.store.Rollback(txid)
		return nil, err
	}

	if err = f.store.Commit(txid); err != nil {
		f.store.Rollback(txid)
		return nil, err
	}

	return j, nil
}
