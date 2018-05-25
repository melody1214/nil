package object

import (
	"fmt"
	"io"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

type endec struct {
	chunkPool *chunkPool
	store     Repository
	cmapAPI   cmap.SlaveAPI
}

func newEndec(cmapAPI cmap.SlaveAPI, p *chunkPool, s Repository) (*endec, error) {
	if cmapAPI == nil || p == nil || s == nil {
		return nil, fmt.Errorf("invalid arguments")
	}

	return &endec{
		chunkPool: p,
		store:     s,
		cmapAPI:   cmapAPI,
	}, nil
}

func (e *endec) Run() {
	ctxLogger := mlog.GetMethodLogger(logger, "endec.Run")

	checkEncodingJobTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-checkEncodingJobTimer.C:
			go func() {
				err := e.checkRoutine()
				if err != nil {
					ctxLogger.Error(err)
				}
			}()
		}
	}
}

// checkRoutine try to encode if there is some waiting chunks for encoded.
func (e *endec) checkRoutine() error {
	ctxLogger := mlog.GetMethodLogger(logger, "endec.checkRoutine")

	c, exist := e.chunkPool.GetNeedEncodingChunk()
	if exist == false {
		// There is no chunk that waiting for encoded.
		return nil
	}

	if isSystemLoadHigh() {
		ctxLogger.Info("current system load is too high to encode chunks. yield cpu for other jobs")
		e.chunkPool.EncodingFailed(c.id)
		return nil
	}

	if err := e.genLocalParity(c); err != nil {
		e.chunkPool.EncodingFailed(c.id)
		return errors.Wrapf(err, "failed to generate local parity for chunk: %+v", c)
	}

	return nil
}

// isSystemLoadHigh checks the current system load and returns true or false.
// TODO: implementation.
func isSystemLoadHigh() bool {
	return false
}

// genLocalParity manages generating local parity job.
func (e *endec) genLocalParity(c chunk) error {
	egID, err := strconv.ParseInt(string(c.encodingGroup), 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to convert encoding group id to cmap id")
	}

	cmapVer := e.cmapAPI.GetLatestCMapVersion()
	eg, err := e.cmapAPI.SearchCall().EncGrp().ID(cmap.ID(egID)).Do()
	if err != nil {
		return errors.Wrap(err, "failed to find encoding group")
	}

	if eg.Stat != cmap.EGAlive {
		return fmt.Errorf("give up to generate local parity because target encoding group is not alive")
	}

	stopC := make(chan interface{}, 1)
	defer close(stopC)

	timeoutC := time.After(5 * time.Minute)
	encodingC := e._genLocalParity(c, stopC)
	cmapChangedC := e.cmapAPI.GetUpdatedNoti(cmapVer)
	for {
		select {
		case err = <-encodingC:
			if err != nil {
				return errors.Wrap(err, "error occured in calculating local parity")
			}
			return nil
		case <-cmapChangedC:
			cmapVer = e.cmapAPI.GetLatestCMapVersion()
			newEg, err := e.cmapAPI.SearchCall().EncGrp().ID(cmap.ID(egID)).Do()
			if err != nil {
				stopC <- nil
				return errors.Wrap(err, "failed to find encoding group")
			}

			if newEg.Stat != cmap.EGAlive {
				stopC <- nil
				return fmt.Errorf("encoding group status has changed to not alive while in encoding")
			}

			for idx, v := range eg.Vols {
				if newEg.Vols[idx] != v {
					stopC <- nil
					return fmt.Errorf("encoding group volume members has changed while in encoding")
				}
			}

			eg = newEg
		case <-timeoutC:
			stopC <- nil
			return fmt.Errorf("timeout")
		}
	}
}

// _genLocalParity generates the local parity by xoring the same encoding group chunks.
func (e *endec) _genLocalParity(c chunk, stop <-chan interface{}) <-chan error {
	notiC := make(chan error)

	go func(ret chan error, stop <-chan interface{}) {
		if err := e.truncateAllChunks(c); err != nil {
			notiC <- errors.Wrap(err, "failed to truncate chunk")
			return
		}

		prArr := make([]*io.PipeReader, e.chunkPool.shardSize)
		pwArr := make([]*io.PipeWriter, e.chunkPool.shardSize)
		bufArr := make([][]byte, e.chunkPool.shardSize)
		readReqArr := make([]*repository.Request, e.chunkPool.shardSize)

		const bufSize int = 1
		for i := int64(0); i < e.chunkPool.shardSize; i++ {
			prArr[i], pwArr[i] = io.Pipe()
			defer pwArr[i].Close()
			defer prArr[i].Close()

			bufArr[i] = make([]byte, bufSize)

			readReqArr[i] = &repository.Request{
				Op:     repository.ReadAll,
				Vol:    string(c.volume),
				LocGid: string(c.encodingGroup),
				Cid:    string(c.id) + "_" + strconv.FormatInt(i+1, 10),
				Osize:  c.size,
				Out:    pwArr[i],
			}
		}

		select {
		// Stop signal sent from manager.
		case <-stop:
			ret <- nil
			return
		default:
			break
		}

		for i := int64(0); i < e.chunkPool.shardSize; i++ {
			e.store.Push(readReqArr[i])
			go func(readReq *repository.Request, idx int64) {
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

		parityWriteReq := &repository.Request{
			Op:     repository.WriteAll,
			Vol:    string(c.volume),
			LocGid: string(c.encodingGroup),
			Oid:    string(c.id),
			Cid:    string(c.id),
			Osize:  c.size,
			In:     parityReader,
		}
		e.store.Push(parityWriteReq)

		for n := int64(0); n < c.size; n++ {
			parityBuf[0] = 0x00
			for i := int64(0); i < e.chunkPool.shardSize; i++ {
				if _, err := prArr[i].Read(bufArr[i]); err != nil {
					ret <- errors.Wrap(err, "failed to read chunk")
					return
				}

				parityBuf[0] = parityBuf[0] ^ bufArr[i][0]
			}

			_, err := parityWriter.Write(parityBuf)
			if err != nil {
				ret <- errors.Wrap(err, "failed to write a xored byte into parity chunk")
				return
			}

			select {
			// Stop signal sent from manager.
			case <-stop:
				for i := int64(0); i < e.chunkPool.shardSize; i++ {
					pwArr[i].CloseWithError(fmt.Errorf("receive stop encoding signal from manager"))
					prArr[i].CloseWithError(fmt.Errorf("receive stop encoding signal from manager"))
				}
				deleteParityReq := &repository.Request{
					Op:     repository.DeleteReal,
					Vol:    string(c.volume),
					LocGid: string(c.encodingGroup),
					Cid:    string(c.id),
					Oid:    string(c.id),
				}
				e.store.Push(deleteParityReq)
				ret <- nil
				return
			default:
				break
			}
		}

		err := parityWriteReq.Wait()
		if err != nil {
			ret <- errors.Wrap(err, "failed to write parity chunk")
			return
		}

		// TODO: delete chunk.
		deleteReqArr := make([]*repository.Request, e.chunkPool.shardSize)
		for i := int64(0); i < e.chunkPool.shardSize; i++ {
			deleteReqArr[i] = &repository.Request{
				Op:     repository.DeleteReal,
				Vol:    string(c.volume),
				LocGid: string(c.encodingGroup),
				Cid:    string(c.id) + "_" + strconv.FormatInt(i+1, 10),
				Oid:    string(c.id) + "_" + strconv.FormatInt(i+1, 10),
			}

			e.store.Push(deleteReqArr[i])
		}

		if err := e.renameToL(c); err != nil {
			fmt.Printf("%+v\n\n", err)
		}

		notiC <- nil
	}(notiC, stop)

	return notiC
}

func (e *endec) renameToL(c chunk) error {
	egID, _ := strconv.ParseInt(string(c.encodingGroup), 10, 64)

	call := e.cmapAPI.SearchCall()
	eg, err := call.EncGrp().ID(cmap.ID(egID)).Do()
	if err != nil {
		return errors.Wrap(err, "failed to find such encoding group")
	}

	vols := make([]cmap.Volume, len(eg.Vols))
	for i := 0; i < len(eg.Vols); i++ {
		v, err := call.Volume().ID(eg.Vols[i]).Do()
		if err != nil {
			return errors.Wrap(err, "failed to find such volume")
		}
		vols[i] = v
	}

	nodes := make([]cmap.Node, len(eg.Vols))
	for i := 0; i < len(eg.Vols); i++ {
		n, err := call.Node().ID(vols[i].Node).Do()
		if err != nil {
			return errors.Wrap(err, "failed to find such volume")
		}
		nodes[i] = n
	}

	for i, n := range nodes {
		conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			return errors.Wrap(err, "failed to dial")
		}
		defer conn.Close()

		req := &nilrpc.DGERenameChunkRequest{
			Vol:      vols[i].ID.String(),
			EncGrp:   string(c.encodingGroup),
			OldChunk: string(c.id),
			NewChunk: "L_" + string(c.id),
		}
		res := &nilrpc.DGERenameChunkResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.DsGencodingRenameChunk.String(), req, res); err != nil {
			return errors.Wrap(err, "failed to truncate remote chunk")
		}
	}

	return nil
}

func (e *endec) truncateAllChunks(c chunk) error {
	// Truncate locals.
	truncateReq := &repository.Request{
		Op:     repository.Write,
		Vol:    string(c.volume),
		LocGid: string(c.encodingGroup),
		Oid:    "fake, just for truncating",
		Osize:  c.size,
		In:     &io.PipeReader{},
	}
	for i := int64(0); i < e.chunkPool.shardSize; i++ {
		truncateReq.Cid = string(c.id) + "_" + strconv.FormatInt(i+1, 10)
		if err := e.store.Push(truncateReq); err != nil {
			return err
		}
		if err := truncateReq.Wait(); err == nil {
			return fmt.Errorf("truncate request returns no error")
		} else if err.Error() != "truncated" {
			return err
		}
	}

	call := e.cmapAPI.SearchCall()

	// Truncate remotes.
	egID, _ := strconv.ParseInt(string(c.encodingGroup), 10, 64)
	eg, err := call.EncGrp().ID(cmap.ID(egID)).Do()
	if err != nil {
		return errors.Wrap(err, "failed to find such encoding group")
	}

	vols := make([]cmap.Volume, len(eg.Vols)-1)
	for i := 1; i < len(eg.Vols); i++ {
		v, err := call.Volume().ID(eg.Vols[i]).Do()
		if err != nil {
			return errors.Wrap(err, "failed to find such volume")
		}
		vols[i-1] = v
	}

	nodes := make([]cmap.Node, len(eg.Vols)-1)
	for i := 0; i < len(eg.Vols)-1; i++ {
		n, err := call.Node().ID(vols[i].Node).Do()
		if err != nil {
			return errors.Wrap(err, "failed to find such volume")
		}
		nodes[i] = n
	}

	for i, n := range nodes {
		conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			return errors.Wrap(err, "failed to dial")
		}

		req := &nilrpc.DGETruncateChunkRequest{
			Vol:    vols[i].ID.String(),
			EncGrp: string(c.encodingGroup),
			Chunk:  string(c.id),
		}
		res := &nilrpc.DGETruncateChunkResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.DsGencodingTruncateChunk.String(), req, res); err != nil {
			return errors.Wrap(err, "failed to truncate remote chunk")
		}

		conn.Close()
	}

	return nil
}
