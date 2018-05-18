package gencoding

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
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

	tbl, err := s.store.Make()
	if err != nil {
		// fmt.Printf("\n\n%v\n\n", err)
		return
	}

	if err := s.fillTbl(tbl); err != nil {
		fmt.Printf("\n\n%v\n\n", err)
		return
	}
}

func (s *service) fillTbl(tbl *Table) error {
	for i := range tbl.RegionIDs {
		regionEndpoint := s.store.RegionEndpoint(tbl.RegionIDs[i])

		req := &nilrpc.MGESelectEncodingGroupRequest{
			TblID: tbl.ID,
		}
		res := &nilrpc.MGESelectEncodingGroupResponse{}

		conn, err := nilrpc.Dial(regionEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			fmt.Printf("\n\n%v\n\n", err)
			return err
		}
		defer conn.Close()

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.MdsGencodingSelectEncodingGroup.String(), req, res); err != nil {
			fmt.Printf("\n\n%v\n\n", err)
			return err
		}
	}
	return nil
}

func (s *service) SelectEncodingGroup(req *nilrpc.MGESelectEncodingGroupRequest, res *nilrpc.MGESelectEncodingGroupResponse) error {
	// TODO: Get Encoding group and rename unencoded chunk to given tbl id.
	m := s.cmapAPI.GetLatestCMap()

	max := 0
	var target *cmap.EncodingGroup
	for i, eg := range m.EncGrps {
		if max < eg.Uenc {
			max = eg.Uenc
			target = &m.EncGrps[i]
		}
	}

	// There is no unencoded chunk.
	if target == nil {
		return fmt.Errorf("no unencoded chunk")
	}

	v, err := s.cmapAPI.SearchCallVolume().ID(target.Vols[0]).Do()
	if err != nil {
		return err
	}

	n, err := s.cmapAPI.SearchCallNode().ID(v.Node).Do()
	if err != nil {
		return err
	}

	conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	prepareReq := &nilrpc.DGEPrepareEncodingRequest{
		Chunk: strconv.FormatInt(req.TblID, 10),
		Vol:   v.ID,
		EG:    target.ID,
	}
	prepareRes := &nilrpc.DGEPrepareEncodingResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingPrepareEncoding.String(), prepareReq, prepareRes); err != nil {
		return errors.Wrap(err, "failed to rename chunk")
	}

	return nil
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
	SelectEncodingGroup(req *nilrpc.MGESelectEncodingGroupRequest, res *nilrpc.MGESelectEncodingGroupResponse) error
}
