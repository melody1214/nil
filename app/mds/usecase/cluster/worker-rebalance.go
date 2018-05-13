package cluster

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Prefix: rb

// rbStart is the first state of rebalancing.
// Verify the worker and job fields.
func (w *worker) rbStart() fsm {
	w.job.ScheduledAt = TimeNow()

	if w.job.State != Run {
		w.job.err = fmt.Errorf("job state is not Run")
		return w.rbFinish
	}

	return w.rbFinish
}

// rbFinish cleanup the job.
func (w *worker) rbFinish() fsm {
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
