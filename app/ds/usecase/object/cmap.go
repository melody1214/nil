package object

import (
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

type encodeGroup struct {
	id                int64
	firstVolID        int64
	firstVolNodeID    int64
	firstVolNodeAddr  string
	secondVolID       int64
	secondVolNodeID   int64
	secondVolNodeAddr string
	thirdVolID        int64
	thirdVolNodeID    int64
	thirdVolNodeAddr  string
	parityVolID       int64
	parityVolNodeID   int64
	parityVolNodeAddr string
}

func (e *encoder) updateGroup() {
	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		mlog.GetLogger().Error(err)
	}
	if m == nil {
		return
	}

	mds, err := m.SearchCall().Type(cmap.MDS).Do()
	if err != nil {
		mlog.GetLogger().Error(err)
	}

	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		mlog.GetLogger().Error(err)
	}
	defer conn.Close()

	vreq := &nilrpc.GetAllVolumeRequest{}
	vres := &nilrpc.GetAllVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.GetAllVolume.String(), vreq, vres); err != nil {
		mlog.GetLogger().Error(err)
	}

	volumeMap := make(map[int64]int64)
	for _, v := range vres.Volumes {
		volumeMap[v.ID] = v.NodeID
	}

	creq := &nilrpc.GetAllChainRequest{}
	cres := &nilrpc.GetAllChainResponse{}

	if err := cli.Call(nilrpc.GetAllChain.String(), creq, cres); err != nil {
		mlog.GetLogger().Error(err)
	}

	for _, c := range cres.Chains {
		g := encodeGroup{
			id:          c.ID,
			firstVolID:  c.FirstVolumeID,
			secondVolID: c.SecondVolumeID,
			thirdVolID:  c.ThirdVolumeID,
			parityVolID: c.ParityVolumeID,
		}

		id, ok := volumeMap[c.FirstVolumeID]
		if !ok {
			mlog.GetLogger().Error("no such first volume")
			continue
		}
		n, err := m.SearchCall().ID(cmap.ID(id)).Do()
		if err != nil {
			mlog.GetLogger().Error(err)
			continue
		}
		g.firstVolNodeID = id
		g.firstVolNodeAddr = n.Addr

		id, ok = volumeMap[c.SecondVolumeID]
		if !ok {
			mlog.GetLogger().Error("no such second volume")
			continue
		}
		n, err = m.SearchCall().ID(cmap.ID(id)).Do()
		if err != nil {
			mlog.GetLogger().Error(err)
			continue
		}
		g.secondVolNodeID = id
		g.secondVolNodeAddr = n.Addr

		id, ok = volumeMap[c.ThirdVolumeID]
		if !ok {
			mlog.GetLogger().Error("no such third volume")
			continue
		}
		n, err = m.SearchCall().ID(cmap.ID(id)).Do()
		if err != nil {
			mlog.GetLogger().Error(err)
			continue
		}
		g.thirdVolNodeID = id
		g.thirdVolNodeAddr = n.Addr

		id, ok = volumeMap[c.ParityVolumeID]
		if !ok {
			mlog.GetLogger().Error("no such volume")
			continue
		}
		n, err = m.SearchCall().ID(cmap.ID(id)).Do()
		if err != nil {
			mlog.GetLogger().Error(err)
			continue
		}
		g.parityVolNodeID = id
		g.parityVolNodeAddr = n.Addr

		e.emap[strconv.FormatInt(g.id, 10)] = g
	}
}