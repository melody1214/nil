package cluster

import (
	"fmt"
	"time"

	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/cmap"
)

// Prefix: rv

// rvStart is the first state of register volume.
// Verify the worker and job fields.
func (w *worker) rvStart() fsm {
	w.job.ScheduledAt = TimeNow()

	if w.job.State != Run {
		w.job.err = fmt.Errorf("job state is not Run")
		return w.rvFinish
	}

	return w.rvInsertDB
}

func (w *worker) rvInsertDB() fsm {
	var private interface{}
	private, w.job.err = w.job.getPrivate()
	if w.job.err != nil {
		return w.rvFinish
	}

	var txid repository.TxID
	txid, w.job.err = w.store.Begin()
	if w.job.err != nil {
		return w.rvFinish
	}

	vols := private.([]cmap.Volume)
	for _, v := range vols {
		v.MaxEG = calcMaxEG(v.Size)

		if w.job.err = w.store.RegisterVolume(txid, &v); w.job.err != nil {
			w.store.Rollback(txid)
			return w.rvFinish
		}
	}

	if w.job.err = w.store.Commit(txid); w.job.err != nil {
		w.store.Rollback(txid)
		return w.rvFinish
	}

	return w.rvUpdateMap
}

func (w *worker) rvUpdateMap() fsm {
	if err := w.updateClusterMap(); err != nil {
		// TODO: handling error
	}

	return w.rvFinish
}

// rvFinish cleanup the job and add send error information to the requester.
func (w *worker) rvFinish() fsm {
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

func calcMaxEG(volumeSize uint64) int {
	if volumeSize <= 0 {
		return 0
	}

	// Test, eg per 100MB,
	return int(volumeSize / 100)
	// return int(volumeSize / 10)
}