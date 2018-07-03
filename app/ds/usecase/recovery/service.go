package recovery

import (
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg     *config.Ds
	cmapAPI cmap.SlaveAPI
	store   Repository
}

// NewService returns a new instance of a recovery service.
func NewService(cfg *config.Ds, cmapAPI cmap.SlaveAPI, store Repository) Service {
	logger = mlog.GetPackageLogger("app/ds/usecase/recovery")

	s := &service{
		cfg:     cfg,
		cmapAPI: cmapAPI,
		store:   store,
	}
	go s.run()

	return s
}

func (s *service) run() {

}

// Service provides handlers for global encoding.
type Service interface {
}
