package server

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/ds/store/volume"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/pkg/errors"
)

// Handler has exposed methods for rpc server.
type Handler struct {
	s *Server
}

func newNilRPCHandler(s *Server) (NilRPCHandler, error) {
	if s == nil {
		return nil, fmt.Errorf("nil server object")
	}

	return &Handler{s: s}, nil
}

// AddVolume adds a new volume with the given device path.
func (h *Handler) AddVolume(req *nilrpc.AddVolumeRequest, res *nilrpc.AddVolumeResponse) error {
	return h.s.handleAddVolume(req, res)
}

func (s *Server) serveNilRPC(l *nilmux.Layer) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go s.nilRPCSrv.ServeConn(conn)
	}
}

// NilRPCHandler is the interface of mds rpc commands.
type NilRPCHandler interface {
	AddVolume(req *nilrpc.AddVolumeRequest, res *nilrpc.AddVolumeResponse) error
}

func (s *Server) handleAddVolume(req *nilrpc.AddVolumeRequest, res *nilrpc.AddVolumeResponse) error {
	lv, err := volume.NewVol(req.DevicePath)
	if err != nil {
		return err
	}

	// 1) Get volume name from mds.
	mds, err := s.swimSrv.GetMDS()
	if err != nil {
		return errors.Wrap(err, "failed to register volume")
	}
	conn, err := nilrpc.Dial(mds, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	registerReq := &nilrpc.RegisterVolumeRequest{
		Ds:     s.cfg.ID,
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
	if err := s.store.AddVolume(lv); err != nil {
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
