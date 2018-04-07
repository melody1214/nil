package clustermap

import (
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Service manages the cluster map.
type Service struct {
	cMap *cmap.Controller
}

// NewService returns a new instance of a cluster map manager.
func NewService(cMap *cmap.Controller) *Service {
	logger = mlog.GetPackageLogger("app/ds/usecase/clustermap")

	return &Service{
		cMap: cMap,
	}
}

// Run starts to update cluster map periodically.
func (s *Service) Run() {
	go updater(s.cMap)
}

func updater(c *cmap.Controller) {
	ctxLogger := mlog.GetFunctionLogger(logger, "Service.updater")

	// Make ticker for routinely rebalancing.
	updateNoti := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-updateNoti.C:
			if err := c.Update(); err != nil {
				ctxLogger.Error(err)
			}
		}
	}
}
