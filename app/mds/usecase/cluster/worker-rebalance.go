package cluster

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

// Use when the job is failed, but need to be rescheduled.
var errNextTime = errors.New("schedule next time")

// Prefix: rb

// rbStart is the first state of rebalancing.
// Verify the worker and job fields.
func (w *worker) rbStart() fsm {
	w.job.ScheduledAt = TimeNow()

	// TODO: handling if some other rebalance job is working,
	// then give up to handle it.

	if w.job.State != Run {
		w.job.err = fmt.Errorf("job state is not Run")
		return w.rbFinish
	}

	return w.rbRebalanceEG
}

// rbRebalanceEG checks if there are some unbalanced encoding groups.
func (w *worker) rbRebalanceEG() fsm {
	// TODO: implement.
	needToMoveEG := false
	if needToMoveEG {
		return w.rbMakeMoveEvent
	}
	return w.rbMakeEG
}

// rbMakeMoveEvent makes proper move events per unbalanced
// encoding groups. The target encoding groups are under
// read-only state until the move event is handled.
func (w *worker) rbMakeMoveEvent() fsm {
	// TODO: implement.
	return w.rbMakeEG
}

// rbMakeEG checks and makes encoding groups if there are some unbalanced
// volumes which can have more encoding groups.
func (w *worker) rbMakeEG() fsm {
	vols, err := w.store.FindAllVolumes(repository.NotTx)
	if err != nil {
		w.job.err = errNextTime
		return w.rbFinish
	}

	// Check need rebalance.
	if needRebalance(vols) == false {
		return w.rbFinish
	}

	if w.job.err = w.rebalanceWithinSameVolumeSpeedGroup(vols); w.job.err != nil {
		return w.rbFinish
	}

	w.job.mapChanged = true
	return w.rbMakeEG
}

// rbFinish cleanup the job.
func (w *worker) rbFinish() fsm {
	if w.job.mapChanged {
		if err := w.updateClusterMap(); err != nil {
			// TODO: handling error
		}
	}

	w.job.FinishedAt = TimeNow()

	if w.job.err == errNextTime {
		w.job.State = Ready
	} else if w.job.err != nil {
		w.job.State = Abort
	} else {
		w.job.State = Done
	}

	if w.job.err != nil {
		w.job.Log = newJobLog(w.job.err.Error())
	}

	if err := w.store.UpdateJob(repository.NotTx, w.job); err != nil {
		// TODO: handling error
	}

	return nil
}

// isUnbalanced checks if the volume has unbalanced encoding group ratio.
func isUnbalanced(v cmap.Volume) bool {
	if v.MaxEG == 0 {
		return false
	}

	if len(v.EncGrps) == 0 {
		return true
	}

	return (len(v.EncGrps)*100)/v.MaxEG < 70
}

// ByFreeChain for sorting volumes by free chain.
type ByFreeChain []cmap.Volume

func (c ByFreeChain) Len() int      { return len(c) }
func (c ByFreeChain) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c ByFreeChain) Less(i, j int) bool {
	return c[i].MaxEG-len(c[i].EncGrps) > c[j].MaxEG-len(c[j].EncGrps)
}

func needRebalance(vols []cmap.Volume) bool {
	for _, v := range vols {
		if isUnbalanced(v) {
			return true
		}
	}

	return false
}

func (w *worker) rebalanceWithinSameVolumeSpeedGroup(vols []cmap.Volume) error {
	ctxLogger := mlog.GetMethodLogger(logger, "worker.rebalanceWithinSameVolumeGroup")

	rebalanced := false
	speedLv := []cmap.VolumeSpeed{cmap.Low, cmap.Mid, cmap.High}
	for _, speed := range speedLv {
		sVols := make([]cmap.Volume, 0)
		for _, v := range vols {
			if v.Speed != speed {
				continue
			}

			sVols = append(sVols, v)
		}

		if err := w.rebalanceVolumeGroup(sVols); err != nil {
			ctxLogger.Error(err)
		} else {
			// rebalanced is set when any volume group is rebalanced.
			rebalanced = true
		}
	}

	if rebalanced == false {
		return fmt.Errorf("there is no rebalanceable volume group")
	}
	return nil
}

func (w *worker) rebalanceVolumeGroup(vols []cmap.Volume) error {
	for _, v := range vols {
		if !isUnbalanced(v) {
			continue
		}

		return w.doRebalance(v, vols)
	}

	return fmt.Errorf("there is no rebalanceable volume set")
}

func (w *worker) doRebalance(target cmap.Volume, group []cmap.Volume) (err error) {
	perm := rand.Perm(len(group))
	shuffledGroup := make([]cmap.Volume, len(group))
	for i, v := range perm {
		shuffledGroup[v] = group[i]
	}

	// N shards + 1 xoring result.
	if len(shuffledGroup) < w.localParityShards+1 {
		return fmt.Errorf("lack of volumes for rebalancing: %+v", group)
	}
	sort.Sort(ByFreeChain(shuffledGroup))

	return w.newEncodingGroup(target, shuffledGroup, w.localParityShards)
}

func (w *worker) newEncodingGroup(primary cmap.Volume, vols []cmap.Volume, shards int) error {
	// Pick volumes from candidates.
	picked := make([]cmap.Volume, 0, shards+1)
	picked = append(picked, primary)
	for i := 0; i < shards; i++ {
		p, err := w.pickOneNewEncodingGroupVolume(picked, vols)
		if err != nil {
			return errors.Wrap(err, "failed to pick new volume")
		}

		picked = append(picked, p)
	}

	// Make encoding group.
	eg := cmap.EncodingGroup{
		Stat: cmap.EGAlive,
		Vols: make([]cmap.ID, shards+1),
	}
	for i, p := range picked {
		eg.Vols[i] = p.ID
	}

	// TODO: prevent duplicated encoding group.

	// Update repository.
	txid, err := w.store.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	err = w.store.MakeNewEncodingGroup(txid, &eg)
	if err != nil {
		w.store.Rollback(txid)
		return errors.Wrap(err, "failed to make new encoding group")
	}
	if err := w.store.Commit(txid); err != nil {
		w.store.Rollback(txid)
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (w *worker) pickOneNewEncodingGroupVolume(picked []cmap.Volume, candidates []cmap.Volume) (cmap.Volume, error) {
	if cap(picked) == len(picked) {
		return cmap.Volume{}, fmt.Errorf("selected encoding group is full")
	}

	for _, c := range candidates {
		if isPicked(c, picked) {
			continue
		}

		if len(c.EncGrps) >= c.MaxEG {
			continue
		}

		if sameFailureDomain(c, picked) {
			continue
		}

		return c, nil
	}

	return cmap.Volume{}, fmt.Errorf("no available volume in the candidates")
}

func isPicked(target cmap.Volume, picked []cmap.Volume) bool {
	for _, v := range picked {
		if v.ID == target.ID {
			return true
		}
	}
	return false
}

func sameFailureDomain(target cmap.Volume, picked []cmap.Volume) bool {
	for _, v := range picked {
		if v.Node == target.Node {
			return true
		}
	}
	return false
}
