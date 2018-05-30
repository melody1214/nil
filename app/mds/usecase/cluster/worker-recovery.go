package cluster

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/pkg/errors"
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

	var faulty cmap.Volume
	var role int
	for i, egv := range eg.Vols {
		v, err := c.Volume().ID(egv.ID).Do()
		if err != nil {
			w.job.err = err
			return w.rcFinish
		}

		if egv.MoveTo != cmap.ID(0) {
			w.job.err = fmt.Errorf("this encoding group is healing by another worker")
			return w.rcFinish
		}

		if v.Stat != cmap.VolActive {
			faulty = v
			role = i
			break
		}
	}

	if faulty.ID == cmap.ID(0) {
		// Cured by others?
		// It's weird. Let's do again.
		return w.rcDiagnose
	}

	failureDomain := make([]cmap.ID, len(eg.Vols))
	for i, vID := range eg.Vols {
		v, err := c.Volume().ID(vID.ID).Do()
		if err != nil {
			w.job.err = err
			return w.rcFinish
		}

		failureDomain[i] = v.Node
	}

	recover, err := w.store.FindReplaceableVolume(repository.NotTx, &eg, &faulty, failureDomain...)
	if err != nil {
		w.job.err = err
		w.job.Log = newJobLog("failed to select " + err.Error())
		return w.rcFinish
	}
	w.job.Log = newJobLog("selected node is " + recover.String())

	var txid repository.TxID
	txid, w.job.err = w.store.Begin()
	if w.job.err != nil {
		w.job.Log = newJobLog(w.job.err.Error())
		return w.rcFinish
	}

	if w.job.err = w.store.SetEGV(txid, eg.ID, role, faulty.ID, recover); w.job.err != nil {
		w.job.Log = newJobLog("failed to set recovery volume")
		w.store.Rollback(txid)
		return w.rcFinish
	}

	if w.job.err = w.store.VolEGIncr(txid, recover); w.job.err != nil {
		w.job.Log = newJobLog(w.job.err.Error())
		w.store.Rollback(txid)
		return w.rcFinish
	}

	if w.job.err = w.store.Commit(txid); w.job.err != nil {
		w.job.Log = newJobLog(w.job.err.Error())
		w.store.Rollback(txid)
		return w.rcFinish
	}

	if w.job.err = w.updateClusterMap(); w.job.err != nil {
		w.job.Log = newJobLog("failed to update cmap with selected recovery volume")
		w.store.SetEGV(repository.NotTx, eg.ID, role, faulty.ID, cmap.ID(0))
		w.store.VolEGDecr(repository.NotTx, recover)
	}

	many := 0
	for _, egv := range eg.Vols {
		v, err := c.Volume().ID(egv.ID).Do()
		if err != nil {
			w.job.err = err
			return w.rcFinish
		}

		if egv.MoveTo != cmap.ID(0) {
			w.job.err = fmt.Errorf("this encoding group is healing by another worker")
			return w.rcFinish
		}

		if v.Stat != cmap.VolActive {
			many = many + 1
		}
	}

	if many == 1 {
		return w.rcLocal
	}
	return w.rcGlobal
}

// rcLocal is the case where only one of the volumes belonging to
// the encoding group has failed and failover is possible with local parity.
func (w *worker) rcLocal() fsm {
	c := w.cmapAPI.SearchCall()
	eg, err := c.EncGrp().ID(w.job.Event.AffectedEG).Do()
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocal: failed to find eg")
		return w.rcFinish
	}

	for _, egv := range eg.Vols {
		if egv.MoveTo == cmap.ID(0) {
			continue
		}

		if egv.ID == eg.LeaderVol() {
			return w.rcLocalPrimary
		}
		return w.rcLocalFollower
	}

	w.job.err = fmt.Errorf("rcLocal: recovery target is gone")
	return w.rcDiagnose
}

func (w *worker) rcLocalPrimary() fsm {
	c := w.cmapAPI.SearchCall()
	eg, err := c.EncGrp().ID(w.job.Event.AffectedEG).Do()
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocalPrimary: failed to find eg")
		return w.rcFinish
	}

	var egv cmap.EGVol
	for _, v := range eg.Vols {
		if v.MoveTo == cmap.ID(0) {
			continue
		}
		egv = v
		break
	}

	if egv.ID == cmap.ID(0) {
		w.job.err = fmt.Errorf("rcLocalPrimary: failed to find target volume")
		return w.rcFinish
	}

	v, err := c.Volume().ID(egv.MoveTo).Do()
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocalPrimary: failed to find move2vol")
		return w.rcFinish
	}

	if v.ID == cmap.ID(0) {
		w.job.err = fmt.Errorf("rcLocalPrimary: failed to find volume")
		return w.rcFinish
	}

	n, err := c.Node().ID(v.Node).Do()
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocalPrimary: failed to find node")
		return w.rcFinish
	}

	// Recover locally encoded chunks.
	listL, err := w.store.FindAllChunks(w.job.Event.AffectedEG, "L")
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocalPrimary: failed to find all L chunks")
		return w.rcFinish
	}
	for _, c := range listL {
		err := w.recoveryChunk(
			n.Addr.String(),
			&nilrpc.DCLRecoveryChunkRequest{
				ChunkID:     strconv.Itoa(c),
				ChunkStatus: "L",
				ChunkEG:     w.job.Event.AffectedEG,
				ChunkVol:    egv.ID,
				TargetVol:   egv.MoveTo,
				Type:        "LocalPrimary",
			},
			&nilrpc.DCLRecoveryChunkResponse{},
		)
		if err != nil {
			w.job.err = errors.Wrapf(err, "rcLocalPrimary: failed to recovery chunk: %d", c)
			return w.rcFinish
		}
	}

	// Recover globally encoded chunks.
	listG, err := w.store.FindAllChunks(w.job.Event.AffectedEG, "G")
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocalPrimary: failed to find all G chunks")
		return w.rcFinish
	}
	_ = listG

	// Recover writing chunks.
	listW, err := w.store.FindAllChunks(w.job.Event.AffectedEG, "W")
	if err != nil {
		w.job.err = errors.Wrap(err, "rcLocalPrimary: failed to find all W chunks")
		return w.rcFinish
	}
	_ = listW

	w.job.Log = newJobLog(fmt.Sprintf("primary: %+v", listL))
	return w.rcFinish
}

func (w *worker) rcLocalFollower() fsm {
	// Recover locally encoded chunks.
	listL, err := w.store.FindAllChunks(w.job.Event.AffectedEG, "L")
	if err != nil {
		w.job.err = err
		w.job.Log = newJobLog(err.Error())
		return w.rcFinish
	}
	_ = listL

	// Recover globally encoded chunks.
	listG, err := w.store.FindAllChunks(w.job.Event.AffectedEG, "G")
	if err != nil {
		w.job.err = err
		w.job.Log = newJobLog(err.Error())
		return w.rcFinish
	}
	_ = listG

	// Recover writing chunks.
	listW, err := w.store.FindAllChunks(w.job.Event.AffectedEG, "W")
	if err != nil {
		w.job.err = err
		w.job.Log = newJobLog(err.Error())
		return w.rcFinish
	}
	_ = listW

	w.job.Log = newJobLog(fmt.Sprintf("follower: %+v", listL))
	return w.rcFinish
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
		w.job.Log = newJobLog(w.job.err.Error())
	} else {
		w.job.State = Done
	}

	if err := w.store.UpdateJob(repository.NotTx, w.job); err != nil {
		// TODO: handling error
	}

	return nil
}

func (w *worker) recoveryChunk(addr string, req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error {
	conn, err := nilrpc.Dial(addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to dial")
	}
	defer conn.Close()

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsClusterRecoveryChunk.String(), req, res); err != nil {
		return err
	}
	defer cli.Close()
	return nil
}
