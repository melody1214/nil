package cluster

import (
	"fmt"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
)

// Prefix: lj

// ljStart is the first state of local join.
// Verify the worker and job fields.
func (w *worker) ljStart() fsm {
	w.job.ScheduledAt = TimeNow()

	if w.job.State != Run {
		w.job.err = fmt.Errorf("job state is not Run")
		return w.ljFinish
	}

	return w.ljInsertDB
}

// ljInsertDB inserts the requested node into the repository.
func (w *worker) ljInsertDB() fsm {
	var private interface{}
	private, w.job.err = w.job.getPrivate()
	if w.job.err != nil {
		return w.ljFinish
	}

	n := private.(cmap.Node)
	if w.job.err = w.store.LocalJoin(n); w.job.err != nil {
		return w.ljFinish
	}

	return w.ljUpdateMap
}

// ljUpdateMap updates the cluster map with the updated db.
func (w *worker) ljUpdateMap() fsm {
	if err := w.updateClusterMap(); err != nil {
		// TODO: handling error
	}

	return w.ljFinish
}

// ljFinish cleanup the job and add send error information to the requester.
func (w *worker) ljFinish() fsm {
	w.job.FinishedAt = TimeNow()

	if w.job.err != nil {
		w.job.State = Abort
	} else {
		w.job.State = Done
	}

	wc, err := w.job.getWaitChannel()
	if err != nil {
		return nil
	}

	select {
	case wc <- w.job.err:
	case <-time.After(2 * time.Second):
		// Todo: handling timeout.
	}

	if err := w.store.UpdateJob(repository.NotTx, w.job); err != nil {
		// TODO: handling error
	}

	return nil
}
