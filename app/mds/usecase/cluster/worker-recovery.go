package cluster

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Prefix: rc

// rcStart is the first state of recovery job.
// Verify the worker and job fields.
func (w *worker) rcStart() fsm {
	w.job.ScheduledAt = TimeNow()

	if w.job.State != Run {
		w.job.err = fmt.Errorf("job state is not Run")
		return w.rcFinish
	}

	return w.rcFinish
}

// rcFinish cleanup the job.
func (w *worker) rcFinish() fsm {
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
