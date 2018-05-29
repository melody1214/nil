package gencoding

import (
	"net/rpc"
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
	jobID := t.JobID

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
	t.Primary.ChunkID = primary.ChunkID

	// Update primary chunk info.
	err = s.setPrimaryChunk(t.Primary, t.JobID)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to update primary chunk info"))
		s.store.SetChunk(t.Primary.ChunkID, t.Primary.EncGrp, "F")
		s.setJobStatus(jobID, Fail)
		return
	}

	// Find the parity group leader ds and ask to start encoding job.
	parity, err := s.cmapAPI.SearchCall().Node().ID(t.Primary.Node).Status(cmap.NodeAlive).Do()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to find such parity node"))
		s.store.SetChunk(t.Primary.ChunkID, t.Primary.EncGrp, "F")
		s.setJobStatus(jobID, Fail)
		return
	}
	conn, err := nilrpc.Dial(parity.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to dial to leader of parity group"))
		s.store.SetChunk(t.Primary.ChunkID, t.Primary.EncGrp, "F")
		s.setJobStatus(jobID, Fail)
		return
	}
	defer conn.Close()

	req := &nilrpc.DGEEncodeRequest{Token: *t}
	res := &nilrpc.DGEEncodeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingEncode.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to call to leader of parity group"))
		s.store.SetChunk(t.Primary.ChunkID, t.Primary.EncGrp, "F")
		s.setJobStatus(jobID, Fail)
		return
	}
	defer cli.Close()
}

func (s *service) setPrimaryChunk(primary token.Unencoded, jobID int64) error {
	leaderAddr := s.store.LeaderEndpoint()
	conn, err := nilrpc.Dial(leaderAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &nilrpc.MGESetPrimaryChunkRequest{
		Primary: primary,
		Job:     jobID,
	}
	res := &nilrpc.MGESetPrimaryChunkResponse{}

	cli := rpc.NewClient(conn)
	defer cli.Close()

	return cli.Call(nilrpc.MdsGencodingSetPrimaryChunk.String(), req, res)
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
	c := s.cmapAPI.SearchCall()

	eg, err := c.EncGrp().MinUenc().Random().Status(cmap.EGAlive).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	v, err := c.Volume().ID(eg.Vols[0].ID).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	n, err := c.Node().ID(v.Node).Do()
	if err != nil {
		return token.Unencoded{}, err
	}

	cid, err := s.store.GetChunk(eg.ID)
	if err != nil {
		return token.Unencoded{}, err
	}

	return token.Unencoded{
		Node:    n.ID,
		Volume:  v.ID,
		EncGrp:  eg.ID,
		ChunkID: cid,
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
