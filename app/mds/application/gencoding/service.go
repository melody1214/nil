package gencoding

import (
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
}

// NewService creates a global encoding service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.SlaveAPI) (Service, error) {
	logger = mlog.GetPackageLogger("app/mds/application/gencoding")

	s := &service{}

	return s, nil
}

// Service is the interface that provides global encoding domain's service
type Service interface {
}
