package rpchandling

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/ds/store/volume"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/pkg/errors"
)

// AddVolume adds a new volume with the given device path.
func (h *Handler) AddVolume(req *nilrpc.AddVolumeRequest, res *nilrpc.AddVolumeResponse) error {
	lv, err := volume.NewVol(req.DevicePath)
	if err != nil {
		return err
	}

	mds, err := h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			log.Error(err)
			return errors.Wrap(err, "failed to register volume")
		}
	}

	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	registerReq := &nilrpc.RegisterVolumeRequest{
		Ds:     h.nodeID,
		Size:   lv.Size,
		Free:   lv.Free,
		Used:   lv.Used,
		Speed:  lv.Speed.String(),
		Status: lv.Status.String(),
	}
	registerRes := &nilrpc.RegisterVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.RegisterVolume.String(), registerReq, registerRes); err != nil {
		return err
	}

	lv.Name = registerRes.ID
	lv.MntPoint = "vol-" + lv.Name

	// 2) Add lv to the store service.
	if err := h.store.AddVolume(lv); err != nil {
		// TODO: remove added volume in the mds.
		return err
	}

	registerReq.ID = lv.Name
	registerReq.Size = lv.Size
	registerReq.Free = lv.Free
	registerReq.Used = lv.Used
	registerReq.Speed = lv.Speed.String()
	registerReq.Status = lv.Status.String()
	if err := cli.Call(nilrpc.RegisterVolume.String(), registerReq, registerRes); err != nil {
		// TODO: remove added volume in the mds and ds.
		return err
	}

	log.Infof("%+v", lv)

	return nil
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *Handler) updateClusterMap() {
	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		log.Error(err)
		return
	}

	h.clusterMap = m
}
