package delivery

import (
	"net"

	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type Service struct {
}

// NewDeliveryService creates a delivery service with necessary dependencies.
func NewDeliveryService(cfg *config.Mds) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("invalid argument")
	}

	log = mlog.GetLogger().WithField("package", "delivery")

	// Resolve gateway address.
	rAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr+":"+cfg.ServerPort)
	if err != nil {
		return nil, errors.Wrap(err, "resolve mds address failed")
	}
	_ = rAddr

	return &Service{}, nil
}
