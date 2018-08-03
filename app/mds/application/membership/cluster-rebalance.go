package membership

import (
	"sync"

	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

var rebalancingLock sync.Mutex

func (s *service) rebalance() {
	rebalancingLock.Lock()
	defer rebalancingLock.Unlock()

	ctxLogger := mlog.GetMethodLogger(logger, "service.rebalance")

	m, err := s.cr.FindLatest()
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	changed := firstPlaceNodeToEncodingMatrixGroup(m)
	if !changed {
		return
	}

	updated, err := s.cr.UpdateWhole(m)
	if err != nil {
		ctxLogger.Error(err)
	}

	s.cmapAPI.UpdateCMap(updated)
}

func firstPlaceNodeToEncodingMatrixGroup(m *cmap.CMap) (changed bool) {
	changed = false

	targets := make([]int, 0)
	for i, n := range m.Nodes {
		if n.Type != cmap.DS {
			continue
		}

		if n.MatrixID != 0 {
			continue
		}

		if n.Size == 0 {
			continue
		}

		if n.Stat != cmap.NodeAlive {
			continue
		}

		targets = append(targets, i)
	}

	for _, target := range targets {
		matrixID, err := pickMostUrgentEncodingMatrix(m)
		if err != nil {
			continue
		}
		m.Nodes[target].MatrixID = matrixID
		changed = true
	}

	return changed
}

func pickMostUrgentEncodingMatrix(m *cmap.CMap) (matrixID int, err error) {
	if len(m.MatrixIDs) == 0 {
		return -1, fmt.Errorf("no available encoding matrices in cluster")
	}

	sizes := make([]uint64, len(m.MatrixIDs))
	for i, id := range m.MatrixIDs {
		for _, n := range m.Nodes {
			if n.MatrixID != id {
				continue
			}

			if n.Stat != cmap.NodeAlive {
				continue
			}

			sizes[i] = sizes[i] + n.Size
		}
	}

	minIdx := 0
	for i, size := range sizes {
		if size < sizes[minIdx] {
			minIdx = i
		}
	}

	return m.MatrixIDs[minIdx], nil
}
