package gencoding

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"

	"github.com/chanyoung/nil/app/mds/usecase/gencoding/token"
	"github.com/chanyoung/nil/pkg/nilrpc"
)

// Status represents the job status of global encoding.
type Status int

const (
	// Ready means the job is ready.
	Ready Status = iota
	// Run means the job is currently working.
	Run
	// Fail means the job is failed.
	Fail
	// Done means the job is finished successfully.
	Done
)

func (s *service) encode() {
	ctxLogger := mlog.GetMethodLogger(logger, "service.encode")

	t, err := s.getEncodingJobToken()
	if err != nil && err.Error() == "no jobs for you" {
		// There is no jobs for me.
		return
	} else if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to get token which is required for global encoding job"))
		return
	}

	// Get job id from the chunk ID.
	jobID, _ := strconv.ParseInt(t.Primary.ChunkID, 10, 64)

	// Fill the parity node information in this region.
	primary, err := s.findPrimary()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to find store parity group"))
		s.setJobStatus(jobID, Fail)
		return
	}
	t.Primary.Node = primary.Node
	t.Primary.Volume = primary.Volume
	t.Primary.EncGrp = primary.EncGrp

	// Find the parity group leader ds and ask to start encoding job.
	parity, err := s.cmapAPI.SearchCall().Node().ID(t.Primary.Node).Status(cmap.NodeAlive).Do()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to find such parity node"))
		s.setJobStatus(jobID, Fail)
		return
	}
	conn, err := nilrpc.Dial(parity.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to dial to leader of parity group"))
		s.setJobStatus(jobID, Fail)
		return
	}
	defer conn.Close()

	req := &nilrpc.DGEEncodeRequest{Token: *t}
	res := &nilrpc.DGEEncodeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingEncode.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to call to leader of parity group"))
		s.setJobStatus(jobID, Fail)
		return
	}
	defer cli.Close()
}

// Get the token for encoding job.
func (s *service) getEncodingJobToken() (*token.Token, error) {
	// If this node is the leader of the global cluster.
	if s.store.AmILeader() {
		return s.store.GetJob(s.cfg.Raft.LocalClusterRegion)
	}

	// If this node isn't the leader, then ask to the leader.
	leaderAddr := s.store.LeaderEndpoint()
	conn, err := nilrpc.Dial(leaderAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &nilrpc.MGEGetEncodingJobRequest{
		Region: s.cfg.Raft.LocalClusterRegion,
	}
	res := &nilrpc.MGEGetEncodingJobResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsGencodingGetEncodingJob.String(), req, res); err != nil {
		return nil, err
	}
	defer cli.Close()

	return &res.Token, nil
}

// findPrimary finds a proper encoding group for acting primary.
func (s *service) findPrimary() (token.Unencoded, error) {
	m := s.cmapAPI.GetLatestCMap()

	min := 999
	var target *cmap.EncodingGroup
	for i, eg := range m.EncGrps {
		if min > eg.Uenc {
			min = eg.Uenc
			target = &m.EncGrps[i]
		}
	}

	// There is no unencoded chunk.
	if target == nil {
		return token.Unencoded{}, fmt.Errorf("no proper primary candidate")
	}

	c := s.cmapAPI.SearchCall()
	v, err := c.Volume().ID(target.Vols[0]).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	n, err := c.Node().ID(v.Node).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	return token.Unencoded{
		Node:   n.ID,
		Volume: v.ID,
		EncGrp: target.ID,
	}, nil
}

func (s *service) setJobStatus(id int64, status Status) error {
	leaderAddr := s.store.LeaderEndpoint()
	conn, err := nilrpc.Dial(leaderAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.MGESetJobStatusRequest{
		ID:     id,
		Status: int(status),
	}
	res := &nilrpc.MGESetJobStatusResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsGencodingSetJobStatus.String(), req, res); err != nil {
		return err
	}
	defer cli.Close()

	return nil
}
