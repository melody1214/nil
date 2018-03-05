package server

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/ds/store/volume"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
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

	// TODO:
	// 1) Get volume name from mds.
	lv.Name = req.DevicePath
	lv.MntPoint = "mnt-" + lv.Name

	// 2) Add lv to the store service.
	if err := s.store.AddVolume(lv); err != nil {
		return err
	}

	log.Infof("%+v", lv)

	return nil
}
