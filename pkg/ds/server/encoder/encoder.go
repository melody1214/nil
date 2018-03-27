package encoder

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chanyoung/nil/pkg/ds/store"
	"github.com/chanyoung/nil/pkg/ds/store/request"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
)

type Encoder struct {
	chunkMap map[string]*chunkMap
	emap     map[string]encodeGroup
	s        store.Service
	q        *queue
	pushCh   chan interface{}
}

type chunkMap struct {
	chunkID string
	seq     int
}

func NewEncoder(s store.Service) *Encoder {
	return &Encoder{
		chunkMap: make(map[string]*chunkMap),
		emap:     make(map[string]encodeGroup),
		s:        s,
		q:        newRequestsQueue(),
		pushCh:   make(chan interface{}, 1),
	}
}

func (e *Encoder) Run() {
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

func (e *Encoder) Push(r *Request) {
	e.q.push(r)
	r.wg.Add(1)
	e.pushCh <- nil
}

func (e *Encoder) doAll() {
	for {
		if r := e.q.pop(); r != nil {
			e.do(r)
			continue
		}

		break
	}
}

func (e *Encoder) do(r *Request) {
	defer r.wg.Done()

	lcid := r.R.Header.Get("Local-Chain-Id")
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

	osize, err := strconv.ParseInt(r.R.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		r.err = err
		return
	}

	parityCID := e.chunkMap[lcid].chunkID + "-" + strconv.Itoa(e.chunkMap[lcid].seq)
	req := &request.Request{
		Op:     request.Write,
		Vol:    r.R.Header.Get("Volume-Id"),
		LocGid: lcid,
		Oid:    strings.Replace(strings.Trim(r.R.RequestURI, "/"), "/", ".", -1),
		Cid:    parityCID,
		Osize:  osize,

		In: r.R.Body,
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
				e.encode(cm, r.R.Header.Get("Volume-Id"), lcid)
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
		mlog.GetLogger().Error(r.err)
		r.err = err
		return
	}

	addr := "https://"
	switch e.chunkMap[lcid].seq {
	case 0:
		addr = addr + lc.firstVolNodeAddr
	case 1:
		addr = addr + lc.secondVolNodeAddr
	case 2:
		addr = addr + lc.thirdVolNodeAddr
	default:
		mlog.GetLogger().Error("no such volume seq")
		r.err = fmt.Errorf("vol seq error")
		return
	}
	addr = addr + r.R.RequestURI

	pipeReader, pipeWriter := io.Pipe()
	copyReq, err := http.NewRequest(r.R.Method, addr, pipeReader)
	if err != nil {
		r.err = err
		return
	}

	var volID string
	switch e.chunkMap[lcid].seq {
	case 0:
		volID = strconv.FormatInt(lc.firstVolID, 10)
	case 1:
		volID = strconv.FormatInt(lc.secondVolID, 10)
	case 2:
		volID = strconv.FormatInt(lc.thirdVolID, 10)
	default:
		mlog.GetLogger().Error("no such volume seq")
		r.err = fmt.Errorf("vol seq error")
		return
	}
	// copyReq.RequestURI = r.R.RequestURI
	copyReq.Header.Add("Local-Chain-Id", lcid)
	copyReq.Header.Add("Volume-Id", volID)
	copyReq.Header.Add("Chunk-Id", e.chunkMap[lcid].chunkID)
	copyReq.ContentLength = osize

	// mlog.GetLogger().Errorf("%+v", *r.R)
	// mlog.GetLogger().Errorf("%+v", *copyReq)

	req = &request.Request{
		Op:     request.Read,
		Vol:    r.R.Header.Get("Volume-Id"),
		LocGid: lcid,
		Oid:    strings.Replace(strings.Trim(r.R.RequestURI, "/"), "/", ".", -1),
		// Cid:   e.chunkMap[lcid].chunkID,
		Cid:   parityCID,
		Osize: osize,
		Out:   pipeWriter,
	}

	r.err = e.s.Push(req)
	if r.err != nil {
		return
	}

	go func(readReq *request.Request) {
		defer pipeWriter.Close()
		err := readReq.Wait()
		if err != nil {
			mlog.GetLogger().Errorf("%+v", err)
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
		mlog.GetLogger().Errorf("%+v", r.err)
		return
	}

	b, _ := ioutil.ReadAll(resp.Body)
	mlog.GetLogger().Infof("%+v", string(b))
}

func (e *Encoder) encode(chunkmap *chunkMap, volID, lgid string) {
	mlog.GetLogger().Info("Start encoding")

	pr1, pw1 := io.Pipe()
	req1 := &request.Request{
		Op:     request.Read,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-0",
		Osize:  30000,
		Out:    pw1,
	}
	e.s.Push(req1)
	go func(readReq *request.Request) {
		defer pw1.Close()
		err := readReq.Wait()
		if err != nil {
			mlog.GetLogger().Errorf("%+v", err)
			return
		}
	}(req1)
	buf1 := make([]byte, 30000)
	pr1.Read(buf1)

	mlog.GetLogger().Info("Encoding step 1")

	pr2, pw2 := io.Pipe()
	req2 := &request.Request{
		Op:     request.Read,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-1",
		Osize:  30000,
		Out:    pw2,
	}
	e.s.Push(req2)
	go func(readReq *request.Request) {
		defer pw2.Close()
		err := readReq.Wait()
		if err != nil {
			mlog.GetLogger().Errorf("%+v", err)
			return
		}
	}(req2)
	buf2 := make([]byte, 30000)
	pr2.Read(buf2)

	mlog.GetLogger().Info("Encoding step 2")

	pr3, pw3 := io.Pipe()
	req3 := &request.Request{
		Op:     request.Read,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-2",
		Osize:  30000,
		Out:    pw3,
	}
	e.s.Push(req3)
	go func(readReq *request.Request) {
		defer pw2.Close()
		err := readReq.Wait()
		if err != nil {
			mlog.GetLogger().Errorf("%+v", err)
			return
		}
	}(req3)
	buf3 := make([]byte, 30000)
	pr3.Read(buf3)

	mlog.GetLogger().Info("Encoding step 3")

	pr4, pw4 := io.Pipe()
	req4 := &request.Request{
		Op:     request.Write,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID,
		Oid:    chunkmap.chunkID,
		Osize:  30000,
		In:     pr4,
	}
	e.s.Push(req4)

	buf4 := make([]byte, 30000)
	for i := 0; i < 30000; i++ {
		buf4[i] = buf1[i] ^ buf2[i] ^ buf3[i]
	}

	mlog.GetLogger().Info("Encoding step 4")

	_, err := pw4.Write(buf4)
	if err != nil {
		mlog.GetLogger().Errorf("error in pw4: %v", err)
	}
	pw4.Close()

	err = req4.Wait()
	if err != nil {
		mlog.GetLogger().Errorf("error in pw4: %v", err)
	}

	mlog.GetLogger().Info("Encoding step 5")

	req1 = &request.Request{
		Op:     request.Delete,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-0",
		Oid:    chunkmap.chunkID + "-0",
	}
	req2 = &request.Request{
		Op:     request.Delete,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-1",
		Oid:    chunkmap.chunkID + "-1",
	}
	req3 = &request.Request{
		Op:     request.Delete,
		Vol:    volID,
		LocGid: lgid,
		Cid:    chunkmap.chunkID + "-2",
		Oid:    chunkmap.chunkID + "-2",
	}
	e.s.Push(req1)
	e.s.Push(req2)
	e.s.Push(req3)

	mlog.GetLogger().Info("Finish encoding")
}
