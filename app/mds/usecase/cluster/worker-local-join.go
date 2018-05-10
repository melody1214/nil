package cluster

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

// ljStart is the first state of local join.
func (w *worker) ljStart() fsm {
	private, err := w.job.getPrivate()
	if err != nil {
		return nil
	}

	wc, err := w.job.getWaitChannel()
	if err != nil {
		return nil
	}

	n := private.(cmap.Node)
	if err := w.store.LocalJoin(n); err != nil {
		return nil
	}
	wc <- nil

	return nil
}
