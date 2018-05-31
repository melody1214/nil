package cluster

import (
	"io"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"fmt"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

type chunk struct {
	id     string
	prefix string
	eg     cmap.ID
	vol    cmap.ID
}

func (s *service) recoveryLocalPrimary(req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "service.recoveryLocalPrimary")

	m := s.cmapAPI.SearchCall()

	eg, err := m.EncGrp().ID(req.ChunkEG).Do()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to find encoding group"))
		return err
	}

	var e error
	downloaded := make([]*repository.Request, 0)
	var lastIdx int
	for i, egv := range eg.Vols {
		if egv.ID == req.ChunkVol {
			continue
		}

		v, err := m.Volume().ID(egv.ID).Status(cmap.VolActive).Do()
		if err != nil {
			e = errors.Wrap(err, "failed to find volume")
			goto ROLLBACK
		}

		n, err := m.Node().ID(v.Node).Status(cmap.NodeAlive).Do()
		if err != nil {
			e = errors.Wrap(err, "failed to find node")
			goto ROLLBACK
		}

		downReq, err := http.NewRequest(
			"GET",
			"https://"+string(n.Addr)+"/chunk",
			nil,
		)

		downReq.Header.Add("Encoding-Group", req.ChunkEG.String())
		downReq.Header.Add("Volume", v.ID.String())
		if req.ChunkStatus == "L" {
			downReq.Header.Add("Chunk-Name", "L_"+req.ChunkID)
		} else if req.ChunkStatus == "G" {
			downReq.Header.Add("Chunk-Name", "G_"+req.ChunkID)
		} else if req.ChunkStatus == "W" {
			err := s.truncateChunk(n.Addr.String(), v.ID, eg.ID, "W_"+req.ChunkID)
			if err != nil {
				ctxLogger.Infof("truncate fail: %s\n", "W_"+req.ChunkID)
				// No truncateable chunk.
				// Writing is stopped at this shard.
				break
			}
			ctxLogger.Infof("truncate: %s\n", "W_"+req.ChunkID)
			downReq.Header.Add("Chunk-Name", "W_"+req.ChunkID)
		} else {
			e = fmt.Errorf("unknown chunk status")
			goto ROLLBACK
		}

		resp, err := http.DefaultClient.Do(downReq)
		if err != nil {
			e = errors.Wrap(err, "failed to download http client do")
			goto ROLLBACK
		}
		defer resp.Body.Close()

		storeReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID + "_" + strconv.Itoa(i),
			In:     resp.Body,
		}
		downloaded = append(downloaded, &repository.Request{
			Op:     repository.DeleteReal,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID + "_" + strconv.Itoa(i),
		})

		if err = s.store.Push(storeReq); err != nil {
			e = errors.Wrap(err, "failed to push to repository")
			goto ROLLBACK
		}
		if err = storeReq.Wait(); err != nil {
			e = errors.Wrap(err, "failed to wait repository")
			goto ROLLBACK
		}

		lastIdx = i
	}

	ctxLogger.Info("download complete")

	if req.ChunkStatus == "W" {
		chunkPoolReq := &nilrpc.DOBSetChunkPoolRequest{
			ID:    req.ChunkID,
			EG:    req.ChunkEG,
			Vol:   req.ChunkVol,
			Shard: lastIdx,
		}
		chunkPoolRes := &nilrpc.DOBSetChunkPoolResponse{}

		conn, err := nilrpc.Dial(s.cfg.ServerAddr+":"+s.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			e = errors.Wrap(err, "failed to dial for setting chunk pool")
			goto ROLLBACK
		}
		defer conn.Close()

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.DsObjectSetChunkPool.String(), chunkPoolReq, chunkPoolRes); err != nil {
			e = errors.Wrap(err, "failed to call for setting chunk pool")
			goto ROLLBACK
		}
		defer cli.Close()

		return nil
	} else {
		ctxLogger.Info("start encoding")
		chunkSize, err := strconv.ParseInt(s.cfg.ChunkSize, 10, 64)
		if err != nil {
			goto ROLLBACK
		}
		prArr := make([]*io.PipeReader, len(downloaded))
		pwArr := make([]*io.PipeWriter, len(downloaded))
		bufArr := make([][]byte, len(downloaded))
		readReqArr := make([]*repository.Request, len(downloaded))

		const bufSize int = 1
		for i := 0; i < len(downloaded); i++ {
			prArr[i], pwArr[i] = io.Pipe()
			defer pwArr[i].Close()
			defer prArr[i].Close()

			bufArr[i] = make([]byte, bufSize)

			readReqArr[i] = &repository.Request{
				Op:     repository.ReadAll,
				Vol:    downloaded[i].Vol,
				LocGid: downloaded[i].LocGid,
				Cid:    downloaded[i].Cid,
				Out:    pwArr[i],
			}
		}

		for i := 0; i < len(downloaded); i++ {
			s.store.Push(readReqArr[i])
			go func(readReq *repository.Request, idx int) {
				defer pwArr[idx].Close()
				err := readReq.Wait()
				if err != nil {
					return
				}
			}(readReqArr[i], i)
		}

		parityReader, parityWriter := io.Pipe()
		defer parityWriter.Close()
		defer parityReader.Close()
		parityBuf := make([]byte, bufSize)

		ctxLogger.Info("start parity writing")
		parityWriteReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID,
			In:     parityReader,
		}
		s.store.Push(parityWriteReq)

		ctxLogger.Info("start read and xoring")
		for n := int64(0); n < chunkSize; n++ {
			parityBuf[0] = 0x00
			for i := 0; i < len(downloaded); i++ {
				if _, err := prArr[i].Read(bufArr[i]); err != nil {
					e = fmt.Errorf("failed in xoring")
					goto ROLLBACK
				}

				parityBuf[0] = parityBuf[0] ^ bufArr[i][0]
			}

			_, err := parityWriter.Write(parityBuf)
			if err != nil {
				e = errors.Wrap(err, "failed to write a xored byte into parity chunk")
				goto ROLLBACK
			}
		}

		ctxLogger.Info("wait parity writing")
		err = parityWriteReq.Wait()
		if err != nil {
			e = errors.Wrap(err, "failed to write parity chunk")
			goto ROLLBACK
		}
	}

	goto FINISH

ROLLBACK:
	ctxLogger.Info("rollback")
	s.store.Push(
		&repository.Request{
			Op:     repository.DeleteReal,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID,
		},
	)
	ctxLogger.Error(e)

FINISH:
	ctxLogger.Info("done")
	for _, delete := range downloaded {
		s.store.Push(delete)
	}

	return nil
}

func (s *service) recoveryLocalFollower(req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "service.recoveryLocalFollower")

	m := s.cmapAPI.SearchCall()

	eg, err := m.EncGrp().ID(req.ChunkEG).Do()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to find encoding group"))
		return err
	}

	var shard int
	for i, egv := range eg.Vols {
		if egv.ID == req.ChunkVol {
			shard = i
			break
		}
	}

	var e error
	downloaded := make([]*repository.Request, 0)
	for i, egv := range eg.Vols {
		if egv.ID == req.ChunkVol {
			continue
		}

		v, err := m.Volume().ID(egv.ID).Status(cmap.VolActive).Do()
		if err != nil {
			e = errors.Wrap(err, "failed to find volume")
			goto ROLLBACK
		}

		n, err := m.Node().ID(v.Node).Status(cmap.NodeAlive).Do()
		if err != nil {
			e = errors.Wrap(err, "failed to find node")
			goto ROLLBACK
		}

		downReq, err := http.NewRequest(
			"GET",
			"https://"+string(n.Addr)+"/chunk",
			nil,
		)

		downReq.Header.Add("Encoding-Group", req.ChunkEG.String())
		downReq.Header.Add("Volume", v.ID.String())
		if req.ChunkStatus == "L" {
			downReq.Header.Add("Chunk-Name", "L_"+req.ChunkID)
		} else if req.ChunkStatus == "G" {
			downReq.Header.Add("Chunk-Name", "G_"+req.ChunkID)
		} else if req.ChunkStatus == "W" {
			if shard == 0 {
				e = fmt.Errorf("Writing chunk but don't have shard")
				goto ROLLBACK
			}

			if i != 0 {
				e = fmt.Errorf("can't reach here")
				goto ROLLBACK
			}

			ctxLogger.Infof("get writing chunk: %s\n", "W_"+req.ChunkID+"_"+strconv.Itoa(shard))
			downReq.Header.Add("Chunk-Name", "W_"+req.ChunkID)
			downReq.Header.Add("Shard", strconv.Itoa(shard))
			break
		} else {
			e = fmt.Errorf("unknown chunk status")
			goto ROLLBACK
		}

		resp, err := http.DefaultClient.Do(downReq)
		if err != nil {
			e = errors.Wrap(err, "failed to download http client do")
			goto ROLLBACK
		}
		defer resp.Body.Close()

		storeReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID + "_" + strconv.Itoa(i),
			In:     resp.Body,
		}
		downloaded = append(downloaded, &repository.Request{
			Op:     repository.DeleteReal,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID + "_" + strconv.Itoa(i),
		})

		if req.ChunkStatus == "W" {
			storeReq.Cid = req.ChunkStatus + "_" + req.ChunkID
			downloaded[len(downloaded)-1].Cid = req.ChunkStatus + "_" + req.ChunkID
		}

		if err = s.store.Push(storeReq); err != nil {
			e = errors.Wrap(err, "failed to push to repository")
			goto ROLLBACK
		}
		if err = storeReq.Wait(); err != nil {
			e = errors.Wrap(err, "failed to wait repository")
			goto ROLLBACK
		}
	}

	ctxLogger.Info("download complete")

	if req.ChunkStatus == "W" {
		return nil
	} else {
		ctxLogger.Info("start encoding")
		chunkSize, err := strconv.ParseInt(s.cfg.ChunkSize, 10, 64)
		if err != nil {
			goto ROLLBACK
		}
		prArr := make([]*io.PipeReader, len(downloaded))
		pwArr := make([]*io.PipeWriter, len(downloaded))
		bufArr := make([][]byte, len(downloaded))
		readReqArr := make([]*repository.Request, len(downloaded))

		const bufSize int = 1
		for i := 0; i < len(downloaded); i++ {
			prArr[i], pwArr[i] = io.Pipe()
			defer pwArr[i].Close()
			defer prArr[i].Close()

			bufArr[i] = make([]byte, bufSize)

			readReqArr[i] = &repository.Request{
				Op:     repository.ReadAll,
				Vol:    downloaded[i].Vol,
				LocGid: downloaded[i].LocGid,
				Cid:    downloaded[i].Cid,
				Out:    pwArr[i],
			}
		}

		for i := 0; i < len(downloaded); i++ {
			s.store.Push(readReqArr[i])
			go func(readReq *repository.Request, idx int) {
				defer pwArr[idx].Close()
				err := readReq.Wait()
				if err != nil {
					return
				}
			}(readReqArr[i], i)
		}

		parityReader, parityWriter := io.Pipe()
		defer parityWriter.Close()
		defer parityReader.Close()
		parityBuf := make([]byte, bufSize)

		ctxLogger.Info("start parity writing")
		parityWriteReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID,
			In:     parityReader,
		}
		s.store.Push(parityWriteReq)

		ctxLogger.Info("start read and xoring")
		for n := int64(0); n < chunkSize; n++ {
			parityBuf[0] = 0x00
			for i := 0; i < len(downloaded); i++ {
				if _, err := prArr[i].Read(bufArr[i]); err != nil {
					e = fmt.Errorf("failed in xoring")
					goto ROLLBACK
				}

				parityBuf[0] = parityBuf[0] ^ bufArr[i][0]
			}

			_, err := parityWriter.Write(parityBuf)
			if err != nil {
				e = errors.Wrap(err, "failed to write a xored byte into parity chunk")
				goto ROLLBACK
			}
		}

		ctxLogger.Info("wait parity writing")
		err = parityWriteReq.Wait()
		if err != nil {
			e = errors.Wrap(err, "failed to write parity chunk")
			goto ROLLBACK
		}
	}

	goto FINISH

ROLLBACK:
	ctxLogger.Info("rollback")
	s.store.Push(
		&repository.Request{
			Op:     repository.DeleteReal,
			Vol:    req.TargetVol.String(),
			LocGid: req.ChunkEG.String(),
			Cid:    req.ChunkStatus + "_" + req.ChunkID,
		},
	)
	ctxLogger.Error(e)

FINISH:
	ctxLogger.Info("done")
	for _, delete := range downloaded {
		s.store.Push(delete)
	}

	return nil
}

func (s *service) recoveryGlobalPrimary(req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error {
	return nil
}

func (s *service) recoveryGlobalFollower(req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error {
	return nil
}

func (s *service) recoveryByXor(gen chunk, downloaded ...chunk) error {
	return nil
}

func (s *service) truncateChunk(addr string, v, eg cmap.ID, chunkName string) error {
	req := &nilrpc.DGETruncateChunkRequest{
		Vol:    v.String(),
		EncGrp: eg.String(),
		Chunk:  chunkName,
	}
	res := &nilrpc.DGETruncateChunkResponse{}

	conn, err := nilrpc.Dial(addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to dial")
	}
	defer conn.Close()

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsGencodingTruncateChunk.String(), req, res); err != nil {
		return errors.Wrap(err, "failed to truncate chunk")
	}
	defer cli.Close()

	return nil
}
