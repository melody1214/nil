package gencoding

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg     *config.Mds
	cmapAPI cmap.SlaveAPI
	store   Repository
}

// NewService creates a global encoding service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.SlaveAPI, store Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/gencoding")

	return &service{
		cfg:     cfg,
		cmapAPI: cmapAPI,
		store:   store,
	}
}

// GGG stands for generate global encoding group.
// GGG generates the global encoding group with the given regions.
func (s *service) GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error {
	fmt.Println("!!!!!!!!!!!!!!!!!!!!")

	return nil
}

// Service is the interface that provides global encoding domain's service
type Service interface {
	GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error
}
