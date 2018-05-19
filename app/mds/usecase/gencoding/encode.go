package gencoding

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"

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

		req := &nilrpc.MGEGetEncodingJobRequest{
			Region: s.cfg.Raft.LocalClusterRegion,
		}
		res := &nilrpc.MGEGetEncodingJobResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.MdsGencodingGetEncodingJob.String(), req, res); err != nil {
			fmt.Printf("\n\n%+v\n\n", err)
			return
		}
		t = res.Token

		cli.Close()
		conn.Close()
	}

	fmt.Printf("\n\nEncoding Token: %+v\n\n", t)

	primary, err := s.cmapAPI.SearchCallNode().ID(t.Primary.Node).Status(cmap.Alive).Do()
	if err != nil {
		// TODO: error handling.
	}
	conn, err := nilrpc.Dial(primary.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		// TODO: error handling.
		fmt.Printf("\n\n%+v\n\n", err)
		return
	}
	defer conn.Close()

	req := &nilrpc.DGEEncodeRequest{
		Token: t,
	}
	res := &nilrpc.DGEEncodeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingEncode.String(), req, res); err != nil {
		// TODO: error handling.
		fmt.Printf("\n\n%+v\n\n", err)
		return
	}
	defer cli.Close()
}
