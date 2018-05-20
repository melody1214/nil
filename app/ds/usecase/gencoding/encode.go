package gencoding

import (
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	mdsgencoding "github.com/chanyoung/nil/app/mds/usecase/gencoding"
	"github.com/chanyoung/nil/app/mds/usecase/gencoding/token"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

func (s *service) Encode(req *nilrpc.DGEEncodeRequest, res *nilrpc.DGEEncodeResponse) error {
	go s.encode(req.Token)
	return nil
}

func (s *service) encode(t token.Token) {
	ctxLogger := mlog.GetMethodLogger(logger, "service.encode")

	shards := [3]*token.Unencoded{
		&t.First, &t.Second, &t.Third,
	}

	rollbacks := make([]*repository.Request, 0)
	for _, shard := range shards {
		rb, err := s.downloadChunk(shard, &t.Primary)
		rollbacks = append(rollbacks, rb...)
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to download unencoded chunks for global encoding"))
			goto ROLLBACK
		}
	}

	return

ROLLBACK:
	// 1. Removes all downloaded chunks.
	for _, r := range rollbacks {
		s.store.Push(r)
	}

	// 2. Set the job status failed.
	mds, err := s.cmapAPI.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "rollback failed: failed to find alive mds"))
		return
	}

	// Chunk id is same with the job id.
	jobID, err := strconv.ParseInt(t.Primary.ChunkID, 10, 64)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "rollback failed: failed to convert job id"))
		return
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "rollback failed: failed to dial to mds"))
		return
	}
	defer conn.Close()

	req := &nilrpc.MGESetJobStatusRequest{
		ID:     jobID,
		Status: int(mdsgencoding.Fail),
	}
	res := &nilrpc.MGESetJobStatusResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsGencodingSetJobStatus.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "rollback failed"))
		return
	}
	defer cli.Close()

	// rollback := func(reqs []*repository.Request) {
	// 	for _, r := range reqs {
	// 		s.store.Push(r)
	// 	}
	// }

	// localShards, err := strconv.Atoi(s.cfg.LocalParityShards)
	// if err != nil {
	// 	return
	// }

	// rollbackReq := make([]*repository.Request, 0)

	// fmt.Println("Download first chunk")
	// // Downloads the first chunk.
	// for i := 1; i <= localShards; i++ {
	// 	req, err := http.NewRequest(
	// 		"GET",
	// 		"https://"+string(t.First.Region.Endpoint)+"/chunk",
	// 		nil,
	// 	)

	// 	req.Header.Add("Encoding-Group", t.First.EncGrp.String())
	// 	req.Header.Add("Chunk-Name", t.First.ChunkID)
	// 	req.Header.Add("Shard-Number", strconv.Itoa(i))

	// 	resp, err := http.DefaultClient.Do(req)
	// 	if err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}

	// 	storeReq := &repository.Request{
	// 		Op:     repository.WriteAll,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i),
	// 		In:     resp.Body,
	// 	}

	// 	if err = s.store.Push(storeReq); err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}
	// 	if err = storeReq.Wait(); err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}

	// 	rollbackReq = append(rollbackReq, &repository.Request{
	// 		Op:     repository.DeleteReal,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i),
	// 	})
	// }

	// fmt.Println("Download second chunk")
	// // Downloads the second chunk.
	// for i := 1; i <= localShards; i++ {
	// 	req, err := http.NewRequest(
	// 		"GET",
	// 		"https://"+string(t.Second.Region.Endpoint)+"/chunk",
	// 		nil,
	// 	)

	// 	req.Header.Add("Encoding-Group", t.Second.EncGrp.String())
	// 	req.Header.Add("Chunk-Name", t.Second.ChunkID)
	// 	req.Header.Add("Shard-Number", strconv.Itoa(i))

	// 	resp, err := http.DefaultClient.Do(req)
	// 	if err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}

	// 	storeReq := &repository.Request{
	// 		Op:     repository.WriteAll,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i+localShards),
	// 		In:     resp.Body,
	// 	}

	// 	if err = s.store.Push(storeReq); err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}
	// 	if err = storeReq.Wait(); err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}

	// 	rollbackReq = append(rollbackReq, &repository.Request{
	// 		Op:     repository.DeleteReal,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i+localShards),
	// 	})
	// }

	// fmt.Println("Download third chunk")
	// // Downloads the second chunk.
	// for i := 1; i <= localShards; i++ {
	// 	req, err := http.NewRequest(
	// 		"GET",
	// 		"https://"+string(t.Third.Region.Endpoint)+"/chunk",
	// 		nil,
	// 	)

	// 	req.Header.Add("Encoding-Group", t.Third.EncGrp.String())
	// 	req.Header.Add("Chunk-Name", t.Third.ChunkID)
	// 	req.Header.Add("Shard-Number", strconv.Itoa(i))

	// 	resp, err := http.DefaultClient.Do(req)
	// 	if err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}

	// 	storeReq := &repository.Request{
	// 		Op:     repository.WriteAll,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i+(localShards*2)),
	// 		In:     resp.Body,
	// 	}

	// 	if err = s.store.Push(storeReq); err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}
	// 	if err = storeReq.Wait(); err != nil {
	// 		// TODO: fail handling.
	// 		fmt.Printf("\n\n%+v\n\n", err)
	// 		rollback(rollbackReq)
	// 		return
	// 	}

	// 	rollbackReq = append(rollbackReq, &repository.Request{
	// 		Op:     repository.DeleteReal,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i+(localShards*2)),
	// 	})
	// }

	// input := make([]io.Reader, 15)
	// for i := range input {
	// 	r, w := io.Pipe()
	// 	storeReq := &repository.Request{
	// 		Op:     repository.ReadAll,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "E_" + t.Primary.ChunkID + "_" + strconv.Itoa(i+1),
	// 		Out:    w,
	// 	}
	// 	if err := s.store.Push(storeReq); err != nil {
	// 		rollback(rollbackReq)
	// 		return
	// 	}
	// 	input[i] = r
	// }

	// parity := make([]io.Writer, 5)
	// for i := range parity {
	// 	r, w := io.Pipe()
	// 	storeReq := &repository.Request{
	// 		Op:     repository.WriteAll,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "G_" + t.Primary.ChunkID + "_" + strconv.Itoa(i),
	// 		In:     r,
	// 	}
	// 	if err := s.store.Push(storeReq); err != nil {
	// 		rollback(rollbackReq)
	// 		return
	// 	}
	// 	rollbackReq = append(rollbackReq, &repository.Request{
	// 		Op:     repository.DeleteReal,
	// 		Vol:    t.Primary.Volume.String(),
	// 		LocGid: t.Primary.EncGrp.String(),
	// 		Cid:    "G_" + t.Primary.ChunkID + "_" + strconv.Itoa(i),
	// 	})
	// 	parity[i] = w
	// }

	// fmt.Println("1111")
	// enc, err := reedsolomon.NewStream(15, 10)
	// if err != nil {
	// 	fmt.Printf("%+v\n\n", err)
	// 	rollback(rollbackReq)
	// 	return
	// }

	// fmt.Println("2222")
	// err = enc.Encode(input, parity)
	// if err != nil {
	// 	fmt.Printf("%+v\n\n", err)
	// 	rollback(rollbackReq)
	// 	return
	// }
}

func (s *service) downloadChunk(src, dst *token.Unencoded) (rollbacks []*repository.Request, err error) {
	localShards, err := strconv.Atoi(s.cfg.LocalParityShards)
	if err != nil {
		return rollbacks, err
	}

	rollbacks = make([]*repository.Request, 0)
	for i := 0; i < localShards; i++ {
		req, err := http.NewRequest(
			"GET",
			"https://"+string(src.Region.Endpoint)+"/chunk",
			nil,
		)

		req.Header.Add("Encoding-Group", src.EncGrp.String())
		req.Header.Add("Chunk-Name", src.ChunkID)
		req.Header.Add("Shard-Number", strconv.Itoa(i))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return rollbacks, err
		}

		storeReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    dst.Volume.String(),
			LocGid: dst.EncGrp.String(),
			Cid:    "E_" + src.ChunkID + "_" + strconv.Itoa(i),
			In:     resp.Body,
		}

		if err = s.store.Push(storeReq); err != nil {
			return rollbacks, err
		}
		if err = storeReq.Wait(); err != nil {
			return rollbacks, err
		}

		rollbacks = append(rollbacks, &repository.Request{
			Op:     repository.DeleteReal,
			Vol:    src.Volume.String(),
			LocGid: src.EncGrp.String(),
			Cid:    "E_" + src.ChunkID + "_" + strconv.Itoa(i),
		})
	}

	return rollbacks, nil
}

// GetCandidateChunk selects an unencoded chunk from the given encoding group.
// Makes selected chunk ready to encode.
func (s *service) GetCandidateChunk(req *nilrpc.DGEGetCandidateChunkRequest, res *nilrpc.DGEGetCandidateChunkResponse) (err error) {
	res.Chunk, err = s.store.GetNonCodedChunk(req.Vol.String(), req.EG.String())
	return err
}
