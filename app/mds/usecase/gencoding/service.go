package gencoding

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

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
	// Check and create global encoding jobs in every 60 seconds.
	checkTicker := time.NewTicker(60 * time.Second)

	// update unencoded chunks in every 30 seconds.
	updateTicker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-checkTicker.C:
			s.createGlobalEncodingJob()
		case <-updateTicker.C:
			s.checkAndUpdateUnencodedChunk()
		}
	}
}

func (s *service) createGlobalEncodingJob() {
	// Do create global encoding job is only allowed to master.
	if s.store.AmILeader() == false {
		return
	}

	if err := s.store.Make(); err != nil {
		// fmt.Printf("\n\n%v\n\n", err)
	}
}

func (s *service) checkAndUpdateUnencodedChunk() {
	// No leader to send.
	leaderEndpoint := s.store.LeaderEndpoint()
	if leaderEndpoint == "" {
		return
	}

	m := s.cmapAPI.GetLatestCMap()

	max := 0
	for _, eg := range m.EncGrps {
		if max < eg.Uenc {
			max = eg.Uenc
		}
	}

	req := &nilrpc.MGEUpdateUnencodedChunkRequest{
		Region:    s.cfg.Raft.LocalClusterRegion,
		Unencoded: max,
	}
	res := &nilrpc.MGEUpdateUnencodedChunkResponse{}
	if s.store.AmILeader() {
		err := s.UpdateUnencodedChunk(req, res)
		if err != nil {
			fmt.Printf("\n\n%v\n\n", err)
		}
		return
	}

	conn, err := nilrpc.Dial(leaderEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		fmt.Printf("\n\n%v\n\n", err)
		return
	}
	defer conn.Close()

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsGencodingUpdateUnencodedChunk.String(), req, res); err != nil {
		fmt.Printf("\n\n%v\n\n", err)
	}
}

func (s *service) UpdateUnencodedChunk(req *nilrpc.MGEUpdateUnencodedChunkRequest, res *nilrpc.MGEUpdateUnencodedChunkResponse) error {
	err := s.store.UpdateUnencodedChunks(req.Region, req.Unencoded)
	if err != nil {
		fmt.Printf("\n\n%v\n\n", err)
	}
	return err
}

// Service is the interface that provides global encoding domain's service
type Service interface {
	GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error
	UpdateUnencodedChunk(req *nilrpc.MGEUpdateUnencodedChunkRequest, res *nilrpc.MGEUpdateUnencodedChunkResponse) error
}
