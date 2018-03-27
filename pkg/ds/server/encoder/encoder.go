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
			go func() {
				encode(e.chunkMap[lcid])
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

func encode(chunkmap *chunkMap) {

}
