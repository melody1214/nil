package gencoding

import (
	"fmt"
	"log"
	"net/rpc"
	"runtime"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/mds/usecase/gencoding/token"
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
	tokenM  *token.Manager
	cmapAPI cmap.SlaveAPI
	store   Repository
}

// NewService creates a global encoding service with necessary dependencies.
func NewService(cfg *config.Mds, cmapAPI cmap.SlaveAPI, store Repository) (Service, error) {
	logger = mlog.GetPackageLogger("app/mds/usecase/gencoding")

	s := &service{
		cfg:     cfg,
		tokenM:  token.NewManager(),
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

// func (s *service) run() {
// // Check and create global encoding jobs in every 60 seconds.
// checkTicker := time.NewTicker(60 * time.Second)

// // update unencoded chunks in every 30 seconds.
// updateTicker := time.NewTicker(30 * time.Second)
// for {
// 	select {
// 	case <-checkTicker.C:
// 		s.createGlobalEncodingJob()
// 	case <-updateTicker.C:
// 		s.checkAndUpdateUnencodedChunk()
// 	}
// }
// }

// func (s *service) createGlobalEncodingJob() {
// 	// Do create global encoding job is only allowed to master.
// 	if s.store.AmILeader() == false {
// 		return
// 	}

// 	tbl, err := s.store.Make()
// 	if err != nil {
// 		// fmt.Printf("\n\n%v\n\n", err)
// 		return
// 	}

// 	if err := s.fillTbl(tbl); err != nil {
// 		fmt.Printf("\n\n%v\n\n", err)
// 		return
// 	}
// }

// func (s *service) fillTbl(tbl *Table) error {
// 	for i := range tbl.RegionIDs {
// 		regionEndpoint := s.store.RegionEndpoint(tbl.RegionIDs[i])

// 		req := &nilrpc.MGESelectEncodingGroupRequest{
// 			TblID: tbl.ID,
// 		}
// 		res := &nilrpc.MGESelectEncodingGroupResponse{}

// 		conn, err := nilrpc.Dial(regionEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
// 		if err != nil {
// 			fmt.Printf("\n\n%v\n\n", err)
// 			return err
// 		}
// 		defer conn.Close()

// 		cli := rpc.NewClient(conn)
// 		if err := cli.Call(nilrpc.MdsGencodingSelectEncodingGroup.String(), req, res); err != nil {
// 			fmt.Printf("\n\n%v\n\n", err)
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (s *service) SelectEncodingGroup(req *nilrpc.MGESelectEncodingGroupRequest, res *nilrpc.MGESelectEncodingGroupResponse) error {
// 	// TODO: Get Encoding group and rename unencoded chunk to given tbl id.
// 	m := s.cmapAPI.GetLatestCMap()

// 	max := 0
// 	var target *cmap.EncodingGroup
// 	for i, eg := range m.EncGrps {
// 		if max < eg.Uenc {
// 			max = eg.Uenc
// 			target = &m.EncGrps[i]
// 		}
// 	}

// 	// There is no unencoded chunk.
// 	if target == nil {
// 		return fmt.Errorf("no unencoded chunk")
// 	}

// 	v, err := s.cmapAPI.SearchCallVolume().ID(target.Vols[0]).Do()
// 	if err != nil {
// 		return err
// 	}

// 	n, err := s.cmapAPI.SearchCallNode().ID(v.Node).Do()
// 	if err != nil {
// 		return err
// 	}

// 	conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close()

// 	prepareReq := &nilrpc.DGEPrepareEncodingRequest{
// 		Chunk: strconv.FormatInt(req.TblID, 10),
// 		Vol:   v.ID,
// 		EG:    target.ID,
// 	}
// 	prepareRes := &nilrpc.DGEPrepareEncodingResponse{}

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.DsGencodingPrepareEncoding.String(), prepareReq, prepareRes); err != nil {
// 		return errors.Wrap(err, "failed to rename chunk")
// 	}

// 	return nil
// }

// func (s *service) checkAndUpdateUnencodedChunk() {
// 	// No leader to send.
// 	leaderEndpoint := s.store.LeaderEndpoint()
// 	if leaderEndpoint == "" {
// 		return
// 	}

// 	m := s.cmapAPI.GetLatestCMap()

// 	max := 0
// 	for _, eg := range m.EncGrps {
// 		if max < eg.Uenc {
// 			max = eg.Uenc
// 		}
// 	}

// 	req := &nilrpc.MGEUpdateUnencodedChunkRequest{
// 		Region:    s.cfg.Raft.LocalClusterRegion,
// 		Unencoded: max,
// 	}
// 	res := &nilrpc.MGEUpdateUnencodedChunkResponse{}
// 	if s.store.AmILeader() {
// 		err := s.UpdateUnencodedChunk(req, res)
// 		if err != nil {
// 			fmt.Printf("\n\n%v\n\n", err)
// 		}
// 		return
// 	}

// 	conn, err := nilrpc.Dial(leaderEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		fmt.Printf("\n\n%v\n\n", err)
// 		return
// 	}
// 	defer conn.Close()

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.MdsGencodingUpdateUnencodedChunk.String(), req, res); err != nil {
// 		fmt.Printf("\n\n%v\n\n", err)
// 	}
// }

// func (s *service) UpdateUnencodedChunk(req *nilrpc.MGEUpdateUnencodedChunkRequest, res *nilrpc.MGEUpdateUnencodedChunkResponse) error {
// 	err := s.store.UpdateUnencodedChunks(req.Region, req.Unencoded)
// 	if err != nil {
// 		fmt.Printf("\n\n%v\n\n", err)
// 	}
// 	return err
// }

func (s *service) run() {
	issueTokenTicker := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-issueTokenTicker.C:
			if !s.store.AmILeader() {
				break
			}

			leg, err := s.store.GetRoutes(s.store.LeaderEndpoint())
			if err != nil {
				fmt.Printf("\n\n%v\n\n", err)
				break
			}

			token := s.tokenM.NewToken(*leg)
			fmt.Printf("\n\n%v\n\n", token)
			s.sendToken(token)
		}
	}
}

func (s *service) sendToken(t *token.Token) {
	nextRoute := t.Routing.Next()
	// fmt.Printf("\nHead to %s: %s\n", nextRoute.RegionName, nextRoute.Endpoint)

	req := &nilrpc.MGEHandleTokenRequest{
		Token: *t,
	}
	res := &nilrpc.MGEHandleTokenResponse{}

	conn, err := nilrpc.Dial(string(nextRoute.Endpoint), nilrpc.RPCNil, time.Duration(15*time.Second))
	if err != nil {
		// fmt.Printf("\n%+v\n", err)
		return
	}

	cli := rpc.NewClient(conn)
	go func() {
		cli.Call(nilrpc.MdsGencodingHandleToken.String(), req, res)
		cli.Close()
		conn.Close()
	}()
	runtime.Gosched()
}

func (s *service) HandleToken(req *nilrpc.MGEHandleTokenRequest, res *nilrpc.MGEHandleTokenResponse) error {
	fmt.Printf("Here token: %v\n", req.Token)
	// Traverse finish and the token returns.
	if req.Token.Routing.CurrentIdx == len(req.Token.Routing.Stops) {
		return nil
	}

	tkn, err := s.findCandidate()
	if err != nil {
		// Give up, and send to other region.
		// fmt.Printf("\n%+v\n", err)
		s.sendToken(&req.Token)
		return nil
	}

	// fmt.Printf("Token: %v\n", tkn)

	// Try to add our unencoded chunk into the global encoding request token.
	req.Token.Add(tkn)
	s.sendToken(&req.Token)
	return nil
}

// findCandidate finds a candidate chunk for global encoding.
func (s *service) findCandidate() (token.Unencoded, error) {
	m := s.cmapAPI.GetLatestCMap()

	priority := 0
	var target *cmap.EncodingGroup
	for i, eg := range m.EncGrps {
		if priority < eg.Uenc {
			priority = eg.Uenc
			target = &m.EncGrps[i]
		}
	}

	// There is no unencoded chunk.
	if target == nil {
		return token.Unencoded{}, fmt.Errorf("no unencoded chunk")
	}

	v, err := s.cmapAPI.SearchCallVolume().ID(target.Vols[0]).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	n, err := s.cmapAPI.SearchCallNode().ID(v.Node).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	prepareReq := &nilrpc.DGEPrepareEncodingRequest{
		Vol: v.ID,
		EG:  target.ID,
	}
	prepareRes := &nilrpc.DGEPrepareEncodingResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingPrepareEncoding.String(), prepareReq, prepareRes); err != nil {
		return token.Unencoded{}, errors.Wrap(err, "failed to rename chunk")
	}
	if prepareRes.Chunk == "" {
		return token.Unencoded{}, fmt.Errorf("no unencoded chunk")
	}

	return token.Unencoded{
		Region:   token.Endpoint(s.cfg.Raft.LocalClusterAddr),
		Node:     n.ID,
		Volume:   v.ID,
		EncGrp:   target.ID,
		ChunkID:  prepareRes.Chunk,
		Priority: priority,
	}, nil
}

// Service is the interface that provides global encoding domain's service
type Service interface {
	HandleToken(req *nilrpc.MGEHandleTokenRequest, res *nilrpc.MGEHandleTokenResponse) error
	GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error
	// UpdateUnencodedChunk(req *nilrpc.MGEUpdateUnencodedChunkRequest, res *nilrpc.MGEUpdateUnencodedChunkResponse) error
	// SelectEncodingGroup(req *nilrpc.MGESelectEncodingGroupRequest, res *nilrpc.MGESelectEncodingGroupResponse) error
}
