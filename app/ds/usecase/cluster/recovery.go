package cluster

import (
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
		if req.ChunkStatus == "L" {
			downReq.Header.Add("Chunk-Name", "L_"+req.ChunkID)
		} else if req.ChunkStatus == "G" {
			downReq.Header.Add("Chunk-Name", "G_"+req.ChunkID)
		} else if req.ChunkStatus == "W" {
			err := s.truncateChunk(n.Addr.String(), v.ID, eg.ID, "W_"+req.ChunkID)
			if err != nil {
				// No truncateable chunk.
				// Writing is stopped at this shard.
				break
			}
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
	}

	ctxLogger.Infof("downloaded for recovery: %+v", downloaded)

	return nil

ROLLBACK:
	ctxLogger.Error(e)
	for _, delete := range downloaded {
		s.store.Push(delete)
	}

	return e
}

func (s *service) recoveryLocalFollower(req *nilrpc.DCLRecoveryChunkRequest, res *nilrpc.DCLRecoveryChunkResponse) error {
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
