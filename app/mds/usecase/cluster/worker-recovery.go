package cluster

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
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

	return w.rcDiagnose
}

// rcDiagnose decide the recovery process would be local or global.
func (w *worker) rcDiagnose() fsm {
	c := w.cmapAPI.SearchCall()
	eg, err := c.EncGrp().ID(w.job.Event.AffectedEG).Do()
	if err != nil {
		w.job.err = err
		return w.rcFinish
	}

	many := 0
	for _, vID := range eg.Vols {
		v, err := c.Volume().ID(vID).Do()
		if err != nil {
			w.job.err = err
			return w.rcFinish
		}

		if v.Stat != cmap.VolActive {
			many = many + 1
		}
	}

	switch many {
	case 0:
		return w.rcFinish
	case 1:
		return w.rcLocal
	default:
		return w.rcGlobal
	}
}

// rcLocal is the case where only one of the volumes belonging to
// the encoding group has failed and failover is possible with local parity.
func (w *worker) rcLocal() fsm {
	c := w.cmapAPI.SearchCall()
	eg, err := c.EncGrp().ID(w.job.Event.AffectedEG).Do()
	if err != nil {
		w.job.err = err
		return w.rcFinish
	}

	var faulty *cmap.Volume
	for _, vID := range eg.Vols {
		v, err := c.Volume().ID(vID).Do()
		if err != nil {
			w.job.err = err
			return w.rcFinish
		}

		if v.Stat != cmap.VolActive {
			faulty = &v
			break
		}
	}

	if faulty == nil {
		// Cured by others?
		// It's weird. Let's do again.
		return w.rcDiagnose
	}

	failureDomain := make([]cmap.ID, len(eg.Vols))
	for i, vID := range eg.Vols {
		v, err := c.Volume().ID(vID).Do()
		if err != nil {
			w.job.err = err
			return w.rcFinish
		}

		failureDomain[i] = v.Node
	}

	recover, err := w.store.FindReplaceableVolume(repository.NotTx, &eg, faulty, failureDomain...)
	if err != nil {
		w.job.err = err
		w.job.Log = newJobLog("failed to select " + err.Error())
		return w.rcFinish
	}
	w.job.Log = newJobLog("selected node is " + recover.String())
	return w.rcFinish

	// return w.rcDiagnose
}

// rcGlobal is the case where more than one of the volumes belonging to
// the encoding group has failed and failover is possible with global parity.
func (w *worker) rcGlobal() fsm {
	return w.rcDiagnose
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
