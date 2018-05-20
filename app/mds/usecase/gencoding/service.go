package gencoding

import (
	"fmt"
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

func (s *service) run() {
	ctxLogger := mlog.GetMethodLogger(logger, "service.run")

	issueTokenTicker := time.NewTicker(60 * time.Second)
	encodeTicker := time.NewTicker(60 * time.Second)
	gcTicker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-issueTokenTicker.C:
			// Only global cluster leader can issue encoding tokens.
			if !s.store.AmILeader() {
				break
			}

			// Get tokens routing legs.
			leg, err := s.store.GetRoutes(s.store.LeaderEndpoint())
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to get routes of encoding token"))
				break
			}

			// Make a token and send to the next region.
			token := s.tokenM.NewToken(*leg)
			s.sendToken(token)

		case <-encodeTicker.C:
			// Check and encode if there are some jobs assigned this region.
			s.encode()

		case <-gcTicker.C:
			s.garbageCollect()
		}
	}
}

func (s *service) sendToken(t *token.Token) {
	ctxLogger := mlog.GetMethodLogger(logger, "service.sendToken")

	// Get next region to send token.
	nextRoute := t.Routing.Next()

	req := &nilrpc.MGEHandleTokenRequest{Token: *t}
	res := &nilrpc.MGEHandleTokenResponse{}

	conn, err := nilrpc.Dial(string(nextRoute.Endpoint), nilrpc.RPCNil, time.Duration(15*time.Second))
	if err != nil {
		// Give up to send the token to the next region.
		// It is okay because the global cluster master node will cancel the token
		// if it wouldn't come back in right time.
		ctxLogger.Error(errors.Wrapf(err, "dial to the next region failed: %v", nextRoute))
		return
	}

	cli := rpc.NewClient(conn)
	// To prevent the connection keep alive until the token traverses all the regions,
	// use goroutine for sending the token to the next region.
	go func() {
		cli.Call(nilrpc.MdsGencodingHandleToken.String(), req, res)
		cli.Close()
		conn.Close()
	}()
	runtime.Gosched()
}

func (s *service) HandleToken(req *nilrpc.MGEHandleTokenRequest, res *nilrpc.MGEHandleTokenResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "service.HandleToken")

	// Traverse finish and the token returns to the issuer.
	if req.Token.Routing.CurrentIdx == len(req.Token.Routing.Stops) {
		// Make global encoding job with the token.
		s.makeGlobalEncodingJob(req.Token)
		return nil
	}

	// Below are the case if the token stopped is the stopover place; Not issuer.
	tkn, err := s.findCandidate(req.Token.Routing.Current())
	if err != nil && err.Error() == "no unencoded chunk" {
		// There is no candidate for global encoding.
		// Give up, and send to the next region.
		s.sendToken(&req.Token)
		return nil
	} else if err != nil {
		// Internal error is occured, log it.
		// Give up, and send to the next region.
		ctxLogger.Error(errors.Wrap(err, "failed to find a global encoding candidate"))
		s.sendToken(&req.Token)
		return nil
	}

	// Try to add our unencoded chunk into the global encoding request token.
	req.Token.Add(tkn)
	s.sendToken(&req.Token)
	return nil
}

// findCandidate finds a candidate chunk for global encoding in current region.
func (s *service) findCandidate(region token.Stop) (*token.Unencoded, error) {
	// Get the lateste map.
	m := s.cmapAPI.GetLatestCMap()

	// Find the encoding group that has the most chunks that are not encoded.
	priority := 0
	var target *cmap.EncodingGroup
	for i, eg := range m.EncGrps {
		if priority < eg.Uenc {
			priority = eg.Uenc
			target = &m.EncGrps[i]
		}
	}

	// Our cluster has no unencoded chunks.
	if target == nil {
		return nil, fmt.Errorf("no unencoded chunk")
	}

	v, err := s.cmapAPI.SearchCallVolume().ID(target.Vols[0]).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to search volume ID: %s", target.Vols[0].String())
	}

	n, err := s.cmapAPI.SearchCallNode().ID(v.Node).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to search node ID: %s", v.Node.String())
	}

	conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial the candidate encoding group's master")
	}
	defer conn.Close()

	getChunkReq := &nilrpc.DGEGetCandidateChunkRequest{
		Vol: v.ID,
		EG:  target.ID,
	}
	getChunkRes := &nilrpc.DGEGetCandidateChunkResponse{}

	// Calling to the our selected encoding group master, to tell me what is the name of unencoded chunk.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingGetCandidateChunk.String(), getChunkReq, getChunkRes); err != nil {
		return nil, errors.Wrap(err, "failed to call the candidate encoding group's master")
	}
	if getChunkRes.Chunk == "" {
		return nil, fmt.Errorf("no unencoded chunk")
	}

	return &token.Unencoded{
		Region:   region,
		Node:     n.ID,
		Volume:   v.ID,
		EncGrp:   target.ID,
		ChunkID:  getChunkRes.Chunk,
		Priority: priority,
	}, nil
}

// makeGlobalEncodingJob makes a global encoding job with the given token information.
func (s *service) makeGlobalEncodingJob(t token.Token) {
	ctxLogger := mlog.GetMethodLogger(logger, "service.makeGlobalEncodingJob")

	// Check is timeout.
	if t.Timeout.Before(time.Now()) {
		ctxLogger.Info("token timeout, give up to make global encoding job")
		return
	}

	// Check the information is enough to make encoding job.
	unencodedChunks := [3]token.Unencoded{t.First, t.Second, t.Third}
	for _, c := range unencodedChunks {
		if c.ChunkID == "" || c.EncGrp == cmap.ID(0) || c.Node == cmap.ID(0) || c.Region.RegionID == 0 {
			return
		}
	}

	// Find a proper primary region for parity chunks.
	p, err := s.findParityRegion(&t)
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	if err = s.store.MakeGlobalEncodingJob(&t, &p); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to make global encoding job"))
		return
	}
}

// findParityRegion finds a region for acting parity chunk store.
func (s *service) findParityRegion(t *token.Token) (token.Unencoded, error) {
	candidates := make([]token.Stop, 0)
	for _, s := range t.Routing.Stops {
		if s.RegionName == t.First.Region.RegionName {
			continue
		}
		if s.RegionName == t.Second.Region.RegionName {
			continue
		}
		if s.RegionName == t.Third.Region.RegionName {
			continue
		}
		candidates = append(candidates, s)
	}

	if len(candidates) == 0 {
		return token.Unencoded{}, fmt.Errorf("no available candidates for parity region")
	}

	i := time.Now().Unix() % int64(len(candidates))
	return token.Unencoded{
		Region: candidates[i],
	}, nil
}

func (s *service) GetEncodingJob(req *nilrpc.MGEGetEncodingJobRequest, res *nilrpc.MGEGetEncodingJobResponse) error {
	t, err := s.store.GetJob(req.Region)
	if err != nil {
		return err
	}

	res.Token = *t
	return nil
}

func (s *service) SetJobStatus(req *nilrpc.MGESetJobStatusRequest, res *nilrpc.MGESetJobStatusResponse) error {
	if !s.store.AmILeader() {
		return s.setJobStatus(req.ID, Status(req.Status))
	}
	return s.store.SetJobStatus(req.ID, Status(req.Status))
}

func (s *service) garbageCollect() {
	// Remove failed or timeouted jobs.
	if !s.store.AmILeader() {
		return
	}

	s.store.RemoveFailedJobs()
}

// Service is the interface that provides global encoding domain's service
type Service interface {
	HandleToken(req *nilrpc.MGEHandleTokenRequest, res *nilrpc.MGEHandleTokenResponse) error
	GGG(req *nilrpc.MGEGGGRequest, res *nilrpc.MGEGGGResponse) error
	GetEncodingJob(req *nilrpc.MGEGetEncodingJobRequest, res *nilrpc.MGEGetEncodingJobResponse) error
	SetJobStatus(req *nilrpc.MGESetJobStatusRequest, res *nilrpc.MGESetJobStatusResponse) error
}
