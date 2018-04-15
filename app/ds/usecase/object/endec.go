package object

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/pkg/errors"
)

type endec struct {
	chunkMap map[string]*chunkMap
	emap     map[string]encodeGroup
	cMap     *cmap.Controller
	s        Repository
	q        *queue
	pushCh   chan interface{}

	// Configurations.
	localParityShards int
}

type chunkMap struct {
	chunkID string
	seq     int
}

func newEncoder(cMap *cmap.Controller, s Repository) (*endec, error) {
	ed := &endec{
		chunkMap: make(map[string]*chunkMap),
		emap:     make(map[string]encodeGroup),
		cMap:     cMap,
		s:        s,
		q:        newRequestsQueue(),
		pushCh:   make(chan interface{}, 1),
	}

	if err := ed.getConfigs(); err != nil {
		return nil, err
	}
	return ed, nil
}

func (e *endec) Run() {
	updateMapNoti := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-e.pushCh:
			e.doAll()
		case <-updateMapNoti.C:
			e.updateGroup()
		}
	}
}

func (e *endec) Push(r *request) {
	e.q.push(r)
	r.wg.Add(1)
	e.pushCh <- nil
}

func (e *endec) doAll() {
	for {
		if r := e.q.pop(); r != nil {
			e.do(r)
			continue
		}

		break
	}
}

func (e *endec) do(r *request) {
	defer r.wg.Done()

	ctxLogger := mlog.GetMethodLogger(logger, "endec.do")

	lcid := r.r.Header.Get("Local-Chain-Id")
	lc, ok := e.emap[lcid]
	if !ok {
		r.err = fmt.Errorf("no such local chain")
		return
	}

	_, ok = e.chunkMap[lcid]
	if !ok {
		e.chunkMap[lcid] = &chunkMap{
			chunkID: uuid.Gen(),
			seq:     0,
		}
	}

	osize, err := strconv.ParseInt(r.r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		r.err = err
		return
	}

	parityCID := e.chunkMap[lcid].chunkID + "-" + strconv.Itoa(e.chunkMap[lcid].seq)
	req := &repository.Request{
		Op:     repository.Write,
		Vol:    r.r.Header.Get("Volume-Id"),
		LocGid: lcid,
		Oid:    strings.Replace(strings.Trim(r.r.RequestURI, "/"), "/", ".", -1),
		Cid:    parityCID,
		Osize:  osize,

		In: r.r.Body,
	}

	r.err = e.s.Push(req)
	if r.err != nil {
		return
	}

	r.err = req.Wait()
	if r.err != nil && r.err.Error() == "chunk full" {
	} else if r.err != nil && r.err.Error() == "truncated" {
		if e.chunkMap[lcid].seq == 2 {
			cm := e.chunkMap[lcid]
			go func() {
				e.encode(cm, r.r.Header.Get("Volume-Id"), lcid)
				// Do encode
			}()

			delete(e.chunkMap, lcid)
		} else {
			e.chunkMap[lcid].seq++
		}
		r.wg.Add(1)
		e.do(r)
		return
	} else if r.err != nil {
		ctxLogger.Error(r.err)
		r.err = err
		return
	}

	addr := "https://"
	chunk, ok := e.chunkMap[lcid]
	if ok == false {
		ctxLogger.Error("no such chunk seq")
		r.err = fmt.Errorf("chunk seq error")
		return
	}
	if len(lc.nodeAddrs) < chunk.seq+1 {
		ctxLogger.Error("no such volume seq")
		r.err = fmt.Errorf("vol seq error")
		return
	}
	addr = addr + lc.nodeAddrs[chunk.seq]

	addr = addr + r.r.RequestURI

	pipeReader, pipeWriter := io.Pipe()
	copyReq, err := http.NewRequest(r.r.Method, addr, pipeReader)
	if err != nil {
		r.err = err
		return
	}

	if len(lc.nodeIDs) < chunk.seq+1 {
		ctxLogger.Error("no such volume seq")
		r.err = fmt.Errorf("vol seq error")
		return
	}
	volID := strconv.FormatInt(lc.nodeIDs[chunk.seq], 10)
	copyReq.Header.Add("Encoding-Group-Volume", volID)

	copyReq.Header.Add("Local-Chain-Id", lcid)
	copyReq.Header.Add("Volume-Id", volID)
	copyReq.Header.Add("Chunk-Id", e.chunkMap[lcid].chunkID)
	copyReq.Header.Add("Md5", r.md5)
	copyReq.ContentLength = osize

	req = &repository.Request{
		Op:     repository.Read,
		Vol:    r.r.Header.Get("Volume-Id"),
		LocGid: lcid,
		Oid:    strings.Replace(strings.Trim(r.r.RequestURI, "/"), "/", ".", -1),
		// Cid:   e.chunkMap[lcid].chunkID,
		Cid:   parityCID,
		Osize: osize,
		Out:   pipeWriter,
	}

	r.err = e.s.Push(req)
	if r.err != nil {
		return
	}

	go func(readReq *repository.Request) {
		defer pipeWriter.Close()
		err := readReq.Wait()
		if err != nil {
			ctxLogger.Errorf("%+v", err)
			return
		}
	}(req)

	var netTransport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
		TLSClientConfig:     security.DefaultTLSConfig(),
		TLSHandshakeTimeout: 5 * time.Second,
	}

	var netClient = &http.Client{
		Timeout:   10 * time.Second,
		Transport: netTransport,
	}

	resp, err := netClient.Do(copyReq)
	if err != nil {
		r.err = err
		ctxLogger.Errorf("%+v", r.err)
		return
	}

	b, _ := ioutil.ReadAll(resp.Body)
	ctxLogger.Infof("%+v", string(b))
}

func (e *endec) encode(chunkmap *chunkMap, volID, lgid string) {
	ctxLogger := mlog.GetMethodLogger(logger, "encode.encode")
	ctxLogger.Info("start encoding")

	pr1, pw1 := io.Pipe()
	req1 := &repository.Request{
		Op:     repository.Read,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-0",
		Osize:  20000,
		Out:    pw1,
	}
	e.s.Push(req1)
	go func(readReq *repository.Request) {
		defer pw1.Close()
		err := readReq.Wait()
		if err != nil {
			ctxLogger.Errorf("%+v", err)
			return
		}
	}(req1)
	buf1 := make([]byte, 20000)
	pr1.Read(buf1)

	pr2, pw2 := io.Pipe()
	req2 := &repository.Request{
		Op:     repository.Read,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-1",
		Osize:  20000,
		Out:    pw2,
	}
	e.s.Push(req2)
	go func(readReq *repository.Request) {
		defer pw2.Close()
		err := readReq.Wait()
		if err != nil {
			ctxLogger.Errorf("%+v", err)
			return
		}
	}(req2)
	buf2 := make([]byte, 20000)
	pr2.Read(buf2)

	pr3, pw3 := io.Pipe()
	req3 := &repository.Request{
		Op:     repository.Read,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-2",
		Osize:  20000,
		Out:    pw3,
	}
	e.s.Push(req3)
	go func(readReq *repository.Request) {
		defer pw2.Close()
		err := readReq.Wait()
		if err != nil {
			ctxLogger.Errorf("%+v", err)
			return
		}
	}(req3)
	buf3 := make([]byte, 20000)
	pr3.Read(buf3)

	pr4, pw4 := io.Pipe()
	req4 := &repository.Request{
		Op:     repository.Write,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID,
		Oid:    chunkmap.chunkID,
		Osize:  20000,
		In:     pr4,
	}
	e.s.Push(req4)

	buf4 := make([]byte, 20000)
	for i := 0; i < 20000; i++ {
		buf4[i] = buf1[i] ^ buf2[i] ^ buf3[i]
	}

	_, err := pw4.Write(buf4)
	if err != nil {
		ctxLogger.Errorf("error in pw4: %+v", err)
	}
	pw4.Close()

	err = req4.Wait()
	if err != nil {
		ctxLogger.Errorf("error in pw4: %+v", err)
	}

	req1 = &repository.Request{
		Op:     repository.Delete,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-0",
		Oid:    chunkmap.chunkID + "-0",
	}
	req2 = &repository.Request{
		Op:     repository.Delete,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-1",
		Oid:    chunkmap.chunkID + "-1",
	}
	req3 = &repository.Request{
		Op:     repository.Delete,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-2",
		Oid:    chunkmap.chunkID + "-2",
	}
	e.s.Push(req1)
	e.s.Push(req2)
	e.s.Push(req3)

	ctxLogger.Info("finish encoding")
}

func (e *endec) getConfigs() error {
	mds, err := e.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		return errors.Wrap(err, "failed to search alive mds")
	}

	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to dial to mds")
	}
	defer conn.Close()

	req := &nilrpc.GetClusterConfigRequest{}
	res := &nilrpc.GetClusterConfigResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsAdminGetClusterConfig.String(), req, res); err != nil {
		return errors.Wrap(err, "failed to rpc call to mds")
	}

	e.localParityShards = res.LocalParityShards
	return nil
}