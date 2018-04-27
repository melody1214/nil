package object

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

type endec struct {
	chunkPool *chunkPool
	store     Repository
	cMap      *cmap.Controller

	// chunkMap map[string]*chunkMap
	// emap     map[string]encodeGroup
	// q        *queue
	// pushCh   chan interface{}

	// // Configurations.
	// localParityShards int
}

// type chunkMap struct {
// 	chunkID string
// 	seq     int
// }

func newEndec(m *cmap.Controller, p *chunkPool, s Repository) (*endec, error) {
	if m == nil || p == nil || s == nil {
		return nil, fmt.Errorf("invalid arguments")
	}

	return &endec{
		chunkPool: p,
		store:     s,
		cMap:      m,
	}, nil
	// ed := &endec{
	// 	chunkMap: make(map[string]*chunkMap),
	// 	emap:     make(map[string]encodeGroup),
	// 	cMap:     cMap,
	// 	s:        s,
	// 	q:        newRequestsQueue(),
	// 	pushCh:   make(chan interface{}, 1),
	// }

	// if err := ed.getConfigs(); err != nil {
	// 	return nil, err
	// }
	// return ed, nil
}

func (e *endec) Run() {
	ctxLogger := mlog.GetMethodLogger(logger, "endec.Run")

	updateMapNoti := time.NewTicker(5 * time.Second)

	for {
		select {
		// case <-e.pushCh:
		// 	e.doAll()
		case <-updateMapNoti.C:
			go func() {
				err := e.checkRoutine()
				if err != nil {
					ctxLogger.Error(err)
				}
			}()
			// 	e.updateGroup()
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

	cmapVer := e.cMap.LatestVersion()
	eg, err := e.cMap.SearchCallEncGrp().ID(cmap.ID(egID)).Do()
	if err != nil {
		return errors.Wrap(err, "failed to find encoding group")
	}

	if eg.Status != cmap.EGAlive {
		return fmt.Errorf("give up to generate local parity because target encoding group is not alive")
	}

	stopC := make(chan interface{}, 1)
	defer close(stopC)

	timeoutC := time.After(5 * time.Minute)
	encodingC := e._genLocalParity(c, stopC)
	cmapChangedC := e.cMap.GetUpdatedNoti(cmapVer)
	for {
		select {
		case err = <-encodingC:
			if err != nil {
				return errors.Wrap(err, "error occured in calculating local parity")
			}
		case <-cmapChangedC:
			cmapVer = e.cMap.LatestVersion()
			newEg, err := e.cMap.SearchCallEncGrp().ID(cmap.ID(egID)).Do()
			if err != nil {
				stopC <- nil
				return errors.Wrap(err, "failed to find encoding group")
			}

			if newEg.Status != cmap.EGAlive {
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

	// TODO: truncate all chunks

	go func(ret chan error, stop <-chan interface{}) {
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
			go func(readReq *repository.Request) {
				defer pwArr[i].Close()
				err := readReq.Wait()
				if err != nil {
					return
				}
			}(readReqArr[i])
		}

		parityReader, parityWriter := io.Pipe()
		defer parityWriter.Close()
		defer parityReader.Close()
		parityBuf := make([]byte, bufSize)

		parityWriteReq := &repository.Request{
			Op:     repository.Write,
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
					Op:     repository.Delete,
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

		deleteReqArr := make([]*repository.Request, e.chunkPool.shardSize)
		for i := int64(0); i < e.chunkPool.shardSize; i++ {
			deleteReqArr[i] = &repository.Request{
				Op:     repository.Delete,
				Vol:    string(c.volume),
				LocGid: string(c.encodingGroup),
				Cid:    string(c.id) + "_" + strconv.FormatInt(i+1, 10),
				Oid:    string(c.id) + "_" + strconv.FormatInt(i+1, 10),
			}

			e.store.Push(deleteReqArr[i])
		}

		notiC <- nil
	}(notiC, stop)

	return notiC
}

// func (e *endec) Push(r *request) {
// 	e.q.push(r)
// 	r.wg.Add(1)
// 	e.pushCh <- nil
// }

// func (e *endec) doAll() {
// 	for {
// 		if r := e.q.pop(); r != nil {
// 			e.do(r)
// 			continue
// 		}

// 		break
// 	}
// }

// func (e *endec) do(r *request) {
// 	defer r.wg.Done()

// 	ctxLogger := mlog.GetMethodLogger(logger, "endec.do")

// 	lcid := r.r.Request().Header.Get("Local-Chain-Id")
// 	lc, ok := e.emap[lcid]
// 	if !ok {
// 		r.err = fmt.Errorf("no such local chain")
// 		return
// 	}

// 	_, ok = e.chunkMap[lcid]
// 	if !ok {
// 		e.chunkMap[lcid] = &chunkMap{
// 			chunkID: uuid.Gen(),
// 			seq:     0,
// 		}
// 	}

// 	osize, err := strconv.ParseInt(r.r.Request().Header.Get("Content-Length"), 10, 64)
// 	if err != nil {
// 		r.err = err
// 		return
// 	}

// 	parityCID := e.chunkMap[lcid].chunkID + "-" + strconv.Itoa(e.chunkMap[lcid].seq)
// 	req := &repository.Request{
// 		Op:     repository.Write,
// 		Vol:    r.r.Request().Header.Get("Volume-Id"),
// 		LocGid: lcid,
// 		Oid:    strings.Replace(strings.Trim(r.r.Request().RequestURI, "/"), "/", ".", -1),
// 		Cid:    parityCID,
// 		Osize:  osize,

// 		In: r.r.Request().Body,
// 	}

// 	r.err = e.s.Push(req)
// 	if r.err != nil {
// 		return
// 	}

// 	r.err = req.Wait()
// 	if r.err != nil && r.err.Error() == "chunk full" {
// 	} else if r.err != nil && r.err.Error() == "truncated" {
// 		if e.chunkMap[lcid].seq == 2 {
// 			cm := e.chunkMap[lcid]
// 			go func() {
// 				e.encode(cm, r.r.Request().Header.Get("Volume-Id"), lcid)
// 				// Do encode
// 			}()

// 			delete(e.chunkMap, lcid)
// 		} else {
// 			e.chunkMap[lcid].seq++
// 		}
// 		r.wg.Add(1)
// 		e.do(r)
// 		return
// 	} else if r.err != nil {
// 		ctxLogger.Error(r.err)
// 		r.err = err
// 		return
// 	}

// 	addr := "https://"
// 	chunk, ok := e.chunkMap[lcid]
// 	if ok == false {
// 		ctxLogger.Error("no such chunk seq")
// 		r.err = fmt.Errorf("chunk seq error")
// 		return
// 	}
// 	if len(lc.nodeAddrs) < chunk.seq+1 {
// 		ctxLogger.Error("no such volume seq")
// 		r.err = fmt.Errorf("vol seq error")
// 		return
// 	}
// 	addr = addr + lc.nodeAddrs[chunk.seq]

// 	addr = addr + r.r.Request().RequestURI

// 	pipeReader, pipeWriter := io.Pipe()

// 	if len(lc.nodeIDs) < chunk.seq+1 {
// 		ctxLogger.Error("no such volume seq")
// 		r.err = fmt.Errorf("vol seq error")
// 		return
// 	}
// 	volID := lc.Vols[chunk.seq].String()

// 	req = &repository.Request{
// 		Op:     repository.Read,
// 		Vol:    r.r.Request().Header.Get("Volume-Id"),
// 		LocGid: lcid,
// 		Oid:    strings.Replace(strings.Trim(r.r.Request().RequestURI, "/"), "/", ".", -1),
// 		// Cid:   e.chunkMap[lcid].chunkID,
// 		Cid:   parityCID,
// 		Osize: osize,
// 		Out:   pipeWriter,
// 	}

// 	r.err = e.s.Push(req)
// 	if r.err != nil {
// 		return
// 	}

// 	go func(readReq *repository.Request) {
// 		defer pipeWriter.Close()
// 		err := readReq.Wait()
// 		if err != nil {
// 			ctxLogger.Errorf("%+v", err)
// 			return
// 		}
// 	}(req)

// 	headers := client.NewHeaders()
// 	headers.SetLocalChainID(lcid)
// 	headers.SetVolumeID(volID)
// 	headers.SetChunkID(e.chunkMap[lcid].chunkID)
// 	headers.SetMD5(r.r.MD5())
// 	copyReq, err := cli.NewRequest(
// 		client.WriteToFollower,
// 		r.r.Request().Method, addr, pipeReader,
// 		headers, osize, cli.WithS3(true),
// 		cli.WithCopyHeaders(r.r.CopyAuthHeader()))
// 	if err != nil {
// 		r.err = err
// 		return
// 	}
// 	resp, err := copyReq.Send()
// 	if err != nil {
// 		r.err = err
// 		ctxLogger.Errorf("%+v", r.err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		r.err = err
// 		ctxLogger.Errorf("%+v", r.err)
// 		return
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		r.err = err
// 		ctxLogger.Errorf("%+v", r.err)
// 		return
// 	}
// 	ctxLogger.Infof("%+v", string(b))
// }

// func (e *endec) encode(chunkmap *chunkMap, volID, lgid string) {
// 	ctxLogger := mlog.GetMethodLogger(logger, "encode.encode")
// 	ctxLogger.Info("start encoding")

// 	pr1, pw1 := io.Pipe()
// 	req1 := &repository.Request{
// 		Op:     repository.ReadAll,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID + "-0",
// 		Osize:  20000,
// 		Out:    pw1,
// 	}
// 	e.s.Push(req1)
// 	go func(readReq *repository.Request) {
// 		defer pw1.Close()
// 		err := readReq.Wait()
// 		if err != nil {
// 			ctxLogger.Errorf("%+v", err)
// 			return
// 		}
// 	}(req1)
// 	buf1 := make([]byte, 20000)
// 	pr1.Read(buf1)

// 	pr2, pw2 := io.Pipe()
// 	req2 := &repository.Request{
// 		Op:     repository.ReadAll,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID + "-1",
// 		Osize:  20000,
// 		Out:    pw2,
// 	}
// 	e.s.Push(req2)
// 	go func(readReq *repository.Request) {
// 		defer pw2.Close()
// 		err := readReq.Wait()
// 		if err != nil {
// 			ctxLogger.Errorf("%+v", err)
// 			return
// 		}
// 	}(req2)
// 	buf2 := make([]byte, 20000)
// 	pr2.Read(buf2)

// 	pr3, pw3 := io.Pipe()
// 	req3 := &repository.Request{
// 		Op:     repository.ReadAll,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID + "-2",
// 		Osize:  20000,
// 		Out:    pw3,
// 	}
// 	e.s.Push(req3)
// 	go func(readReq *repository.Request) {
// 		defer pw2.Close()
// 		err := readReq.Wait()
// 		if err != nil {
// 			ctxLogger.Errorf("%+v", err)
// 			return
// 		}
// 	}(req3)
// 	buf3 := make([]byte, 20000)
// 	pr3.Read(buf3)

// 	pr4, pw4 := io.Pipe()
// 	req4 := &repository.Request{
// 		Op:     repository.Write,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID,
// 		Oid:    chunkmap.chunkID,
// 		Osize:  20000,
// 		In:     pr4,
// 	}
// 	e.s.Push(req4)

// 	buf4 := make([]byte, 20000)
// 	for i := 0; i < 20000; i++ {
// 		buf4[i] = buf1[i] ^ buf2[i] ^ buf3[i]
// 	}

// 	_, err := pw4.Write(buf4)
// 	if err != nil {
// 		ctxLogger.Errorf("error in pw4: %+v", err)
// 	}
// 	pw4.Close()

// 	err = req4.Wait()
// 	if err != nil {
// 		ctxLogger.Errorf("error in pw4: %+v", err)
// 	}

// 	req1 = &repository.Request{
// 		Op:     repository.Delete,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID + "-0",
// 		Oid:    chunkmap.chunkID + "-0",
// 	}
// 	req2 = &repository.Request{
// 		Op:     repository.Delete,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID + "-1",
// 		Oid:    chunkmap.chunkID + "-1",
// 	}
// 	req3 = &repository.Request{
// 		Op:     repository.Delete,
// 		Vol:    volID,
// 		LocGid: lgid,
// 		Cid:    chunkmap.chunkID + "-2",
// 		Oid:    chunkmap.chunkID + "-2",
// 	}
// 	e.s.Push(req1)
// 	e.s.Push(req2)
// 	e.s.Push(req3)

// 	ctxLogger.Info("finish encoding")
// }

// func (e *endec) getConfigs() error {
// 	mds, err := e.cMap.SearchCallNode().Type(cmap.MDS).Status(cmap.Alive).Do()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to search alive mds")
// 	}

// 	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		return errors.Wrap(err, "failed to dial to mds")
// 	}
// 	defer conn.Close()

// 	req := &nilrpc.GetClusterConfigRequest{}
// 	res := &nilrpc.GetClusterConfigResponse{}

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.MdsAdminGetClusterConfig.String(), req, res); err != nil {
// 		return errors.Wrap(err, "failed to rpc call to mds")
// 	}

// 	e.localParityShards = res.LocalParityShards
// 	return nil
// }
