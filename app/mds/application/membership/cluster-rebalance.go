package membership

import (
	"sync"

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

	// Do rebalancing here.
	_ = m
}
