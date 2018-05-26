package cluster

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Prefix: up

// upStart is the first state of update job.
// Verify the worker and job fields.
func (w *worker) upStart() fsm {
	w.job.ScheduledAt = TimeNow()

	if w.job.State != Run {
		w.job.err = fmt.Errorf("job state is not Run")
		return w.ljFinish
	}

	return w.upUpdateMap
}

// upUpdateMap updates the cluster map from the db data.
func (w *worker) upUpdateMap() fsm {
	if err := w.updateClusterMap(); err != nil {
		// TODO: handling error
	}

	return w.upFinish
}

// upFinish cleanup the job.
func (w *worker) upFinish() fsm {
	w.job.FinishedAt = TimeNow()

	if w.job.err != nil {
		w.job.State = Abort
	} else {
		w.job.State = Done
	}

	if err := w.store.UpdateJob(repository.NotTx, w.job); err != nil {
		// TODO: handling error
	}

	return nil
}
