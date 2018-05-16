package gencoding

import (
	"strconv"
	"time"

	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	regions int
	shards  int
	cfg     *config.Mds
	cmapAPI cmap.SlaveAPI
	store   Repository
}

// NewService creates a global encoding service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.SlaveAPI, store Repository) (Service, error) {
	logger = mlog.GetPackageLogger("app/mds/usecase/gencoding")

	s := &service{
		cfg:     cfg,
		cmapAPI: cmapAPI,
		store:   store,
	}

	regions, err := strconv.Atoi(cfg.GlobalParityRegions)
	if err != nil {
		return nil, err
	}
	s.regions = regions

	shards, err := strconv.Atoi(cfg.GlobalParityShards)
	if err != nil {
		return nil, err
	}
	s.shards = shards

	go s.run()

	return s, nil
}

// GGG stands for generate global encoding group.
// GGG generates the global encoding group with the given regions.
func (s *service) GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error {
	if len(req.Regions) != s.regions+1 {
		return fmt.Errorf("invalid region number, required %d number of regions", s.regions)
	}

	return s.store.GenerateGencodingGroup(req.Regions)
}

func (s *service) run() {
	// Check and create global encoding jobs in every 10 seconds.
	checkTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-checkTicker.C:
			s.createGlobalEncodingJob()
		}
	}
}

func (s *service) createGlobalEncodingJob() {
	// Do create global encoding job is only allowed to master.
	if s.store.AmILeader() == false {
		return
	}

	if err := s.store.Make(); err != nil {
		fmt.Printf("\n\n%v\n\n", err)
	}
}

// Service is the interface that provides global encoding domain's service
type Service interface {
	GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error
}
