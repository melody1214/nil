package gencoding

import (
	"io"
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
	"github.com/chanyoung/reedsolomon"
	"github.com/pkg/errors"
)

func (s *service) Encode(req *nilrpc.DGEEncodeRequest, res *nilrpc.DGEEncodeResponse) error {
	go s.encode(req.Token)
	return nil
}

func (s *service) encode(t token.Token) {
	ctxLogger := mlog.GetMethodLogger(logger, "service.encode")

	var enc reedsolomon.StreamEncoder
	input := make([]io.Reader, 0)
	parity := make([]io.Writer, 0)
	downloaded := make([]*repository.Request, 0)
	generated := make([]*repository.Request, 0)
	openedReadStreams := make([]*io.PipeReader, 0)
	openedWriteStreams := make([]*io.PipeWriter, 0)
	bufArr := make([][]byte, 0)
	shards := [3]*token.Unencoded{
		&t.First, &t.Second, &t.Third,
	}
	localShards, err := strconv.Atoi(s.cfg.LocalParityShards)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to convert local parity shards"))
		goto ROLLBACK
	}

	// Download unencoded chunks.
	for _, shard := range shards {
		rb, err := s.downloadChunk(shard, &t.Primary)
		downloaded = append(downloaded, rb...)
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to download unencoded chunks for global encoding"))
			goto ROLLBACK
		}
	}

	// Generate global parities.
	for _, d := range downloaded {
		r, w := io.Pipe()
		readReq := &repository.Request{
			Op:     repository.ReadAll,
			Vol:    d.Vol,
			LocGid: d.LocGid,
			Cid:    d.Cid,
			Out:    w,
		}
		if err := s.store.Push(readReq); err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to push read downloaded chunk request"))
			goto ROLLBACK
		}
		input = append(input, r)
		openedReadStreams = append(openedReadStreams, r)
		openedWriteStreams = append(openedWriteStreams, w)

		go func(rr *repository.Request, pw *io.PipeWriter) {
			err := rr.Wait()
			if err == nil {
				pw.CloseWithError(io.EOF)
			} else {
				pw.CloseWithError(err)
			}
		}(readReq, w)
	}

	for i := 0; i < localShards; i++ {
		r, w := io.Pipe()
		storeReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    t.Primary.Volume.String(),
			LocGid: t.Primary.EncGrp.String(),
			Cid:    "G_" + t.Primary.ChunkID + "_" + strconv.Itoa(i),
			In:     r,
		}
		if err := s.store.Push(storeReq); err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to push write parity chunk request"))
			goto ROLLBACK
		}
		generated = append(generated, &repository.Request{
			Op:     repository.DeleteReal,
			Vol:    t.Primary.Volume.String(),
			LocGid: t.Primary.EncGrp.String(),
			Cid:    "G_" + t.Primary.ChunkID + "_" + strconv.Itoa(i),
		})
		parity = append(parity, w)
		openedReadStreams = append(openedReadStreams, r)
		openedWriteStreams = append(openedWriteStreams, w)
	}

	enc, err = reedsolomon.NewStream(localShards*3, localShards)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to create stream encoder"))
		goto ROLLBACK
	}

	err = enc.Encode(input, parity)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to generate global parities"))
		goto ROLLBACK
	}

	// Remove downloaded chunks.
	for _, d := range downloaded {
		err := s.store.Push(d)
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to delete downloaded chunks"))
			goto ROLLBACK
		}
	}

	// Close streams.
	for _, s := range openedReadStreams {
		s.Close()
	}
	openedReadStreams = make([]*io.PipeReader, 0)

	for _, s := range openedWriteStreams {
		s.Close()
	}
	openedWriteStreams = make([]*io.PipeWriter, 0)

	// Generate local parity.
	for i, g := range generated {
		r, w := io.Pipe()
		openedReadStreams = append(openedReadStreams, r)
		openedWriteStreams = append(openedWriteStreams, w)
		bufArr = append(bufArr, make([]byte, 1))

		storeReq := &repository.Request{
			Op:     repository.ReadAll,
			Vol:    g.Vol,
			LocGid: g.LocGid,
			Cid:    g.Cid,
			Out:    w,
		}
		if err := s.store.Push(storeReq); err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to push read generated chunk request"))
			goto ROLLBACK
		}

		go func(readReq *repository.Request, idx int) {
			defer openedWriteStreams[idx].Close()
			err := readReq.Wait()
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to read generated chunk"))
				return
			}
		}(storeReq, i)
	}

	if true {
		r, w := io.Pipe()
		openedReadStreams = append(openedReadStreams, r)
		openedWriteStreams = append(openedWriteStreams, w)

		parityReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    t.Primary.Volume.String(),
			LocGid: t.Primary.EncGrp.String(),
			Cid:    "G_" + t.Primary.ChunkID,
			In:     r,
		}
		if err := s.store.Push(parityReq); err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to push local parity write request"))
			goto ROLLBACK
		}

		chunkSize, err := strconv.ParseInt(s.cfg.ChunkSize, 10, 64)
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to convert chunk size"))
			goto ROLLBACK
		}

		parityBuf := make([]byte, 1)
		for n := int64(0); n < chunkSize; n++ {
			parityBuf[0] = 0x00
			for i := 0; i < len(bufArr); i++ {
				if _, err := openedReadStreams[i].Read(bufArr[i]); err != nil {
					ctxLogger.Error(errors.Wrap(err, "failed to generate local parity: read fail"))
					goto ROLLBACK
				}

				parityBuf[0] = parityBuf[0] ^ bufArr[i][0]
			}

			_, err := w.Write(parityBuf)
			if err != nil {
				ctxLogger.Error(errors.Wrap(err, "failed to generate local parity: write fail"))
				goto ROLLBACK
			}
		}

		err = parityReq.Wait()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to generate local parity"))
			goto ROLLBACK
		}
	}

	// Move generated parities to other nodes.

	// Register encoded parity to global cluster.

	// Change encoding job status.

	return

ROLLBACK:
	// 1. Removes all downloaded chunks.
	for _, d := range downloaded {
		s.store.Push(d)
	}

	// 2. Removes all generated parity chunks.
	for _, g := range generated {
		s.store.Push(g)
	}

	// 3. Close all opened streams.
	for _, s := range openedReadStreams {
		s.Close()
	}
	for _, s := range openedWriteStreams {
		s.Close()
	}

	// 4. Set the job status failed.
	// Chunk id is same with the job id.
	jobID, err := strconv.ParseInt(t.Primary.ChunkID, 10, 64)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "rollback failed: failed to convert job id"))
		return
	}

	if err := s.setJobStatus(jobID, mdsgencoding.Fail); err != nil {
		ctxLogger.Error(errors.Wrap(err, "rollback failed: failed to change job status"))
		return
	}
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
		defer resp.Body.Close()

		storeReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    dst.Volume.String(),
			LocGid: dst.EncGrp.String(),
			Cid:    "E_" + src.ChunkID + "_" + strconv.Itoa(i),
			In:     resp.Body,
		}
		rollbacks = append(rollbacks, &repository.Request{
			Op:     repository.DeleteReal,
			Vol:    dst.Volume.String(),
			LocGid: dst.EncGrp.String(),
			Cid:    "E_" + src.ChunkID + "_" + strconv.Itoa(i),
		})

		if err = s.store.Push(storeReq); err != nil {
			return rollbacks, err
		}
		if err = storeReq.Wait(); err != nil {
			return rollbacks, err
		}
	}

	return rollbacks, nil
}

// GetCandidateChunk selects an unencoded chunk from the given encoding group.
// Makes selected chunk ready to encode.
func (s *service) GetCandidateChunk(req *nilrpc.DGEGetCandidateChunkRequest, res *nilrpc.DGEGetCandidateChunkResponse) (err error) {
	res.Chunk, err = s.store.GetNonCodedChunk(req.Vol.String(), req.EG.String())
	return err
}

func (s *service) setJobStatus(id int64, status mdsgencoding.Status) error {
	mds, err := s.cmapAPI.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		return err
	}

	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
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
