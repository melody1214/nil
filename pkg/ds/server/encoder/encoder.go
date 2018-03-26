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
)

type Encoder struct {
	emap   map[string]encodeGroup
	s      store.Service
	q      *queue
	pushCh chan interface{}
}

func NewEncoder(s store.Service) *Encoder {
	return &Encoder{
		emap:   make(map[string]encodeGroup),
		s:      s,
		q:      newRequestsQueue(),
		pushCh: make(chan interface{}, 1),
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

	req := &request.Request{
		Op:  request.Write,
		Vol: r.R.Header.Get("Volume-Id"),
		Oid: strings.Replace(strings.Trim(r.R.RequestURI, "/"), "/", ".", -1),

		In: r.R.Body,
	}

	r.err = e.s.Push(req)
	if r.err != nil {
		return
	}

	r.err = req.Wait()
	if r.err != nil {
		return
	}

	pipeReader, pipeWriter := io.Pipe()
	copyReq, err := http.NewRequest(r.R.Method, "https://"+lc.firstVolNodeAddr+r.R.RequestURI, pipeReader)
	if err != nil {
		r.err = err
		return
	}

	// copyReq.RequestURI = r.R.RequestURI
	copyReq.Header.Add("Local-Chain-Id", lcid)
	copyReq.Header.Add("Volume-Id", strconv.FormatInt(lc.firstVolID, 10))

	// mlog.GetLogger().Errorf("%+v", *r.R)
	// mlog.GetLogger().Errorf("%+v", *copyReq)

	req = &request.Request{
		Op:  request.Read,
		Vol: r.R.Header.Get("Volume-Id"),
		Oid: strings.Replace(strings.Trim(r.R.RequestURI, "/"), "/", ".", -1),
		Out: pipeWriter,
	}

	var netTransport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
		TLSClientConfig:     security.DefaultTLSConfig(),
		TLSHandshakeTimeout: 5 * time.Second,
	}

	var netClient = &http.Client{
		Timeout:   10 * time.Second,
		Transport: netTransport,
	}

	r.err = e.s.Push(req)
	if r.err != nil {
		return
	}

	go func() {
		r.err = req.Wait()
		if r.err != nil {
			return
		}
		pipeWriter.Close()
	}()

	resp, err := netClient.Do(copyReq)
	if err != nil {
		r.err = err
		return
	}

	b, _ := ioutil.ReadAll(resp.Body)
	mlog.GetLogger().Infof("%+v", string(b))
}
