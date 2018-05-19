package gencoding

import (
	"fmt"
	"net/rpc"
	"time"

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
	// Get encoding job.
	var t token.Token
	if s.store.AmILeader() {
		tkn, err := s.store.GetJob(s.cfg.Raft.LocalClusterRegion)
		if err != nil {
			fmt.Printf("\n\n%+v\n\n", err)
			return
		}
		t = *tkn
	} else {
		leaderAddr := s.store.LeaderEndpoint()
		conn, err := nilrpc.Dial(leaderAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			fmt.Printf("\n\n%+v\n\n", err)
			return
		}
		defer conn.Close()

		req := &nilrpc.MGEGetEncodingJobRequest{
			Region: s.cfg.Raft.LocalClusterRegion,
		}
		res := &nilrpc.MGEGetEncodingJobResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.MdsGencodingGetEncodingJob.String(), req, res); err != nil {
			fmt.Printf("\n\n%+v\n\n", err)
			return
		}
		cli.Close()
		t = res.Token
	}

	fmt.Printf("\n\nEncoding Token: %+v\n\n", t)
}
