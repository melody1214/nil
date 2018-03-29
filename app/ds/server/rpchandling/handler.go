package rpchandling

import (
	"fmt"

	"github.com/chanyoung/nil/app/ds/store"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// NilRPCHandler is the interface of mds rpc commands.
type NilRPCHandler interface {
	AddVolume(req *nilrpc.AddVolumeRequest, res *nilrpc.AddVolumeResponse) error
}

// TypeBytes returns rpc type bytes which is used to multiplexing.
func TypeBytes() []byte {
	return []byte{
		0x02, // rpcNil
	}
}

// Handler has exposed methods for rpc server.
type Handler struct {
	nodeID     string
	clusterMap *cmap.CMap
	store      store.Service
}

// New returns a new rpc handler.
func New(nodeID string, s store.Service) (NilRPCHandler, error) {
	log = mlog.GetLogger()

	if s == nil {
		return nil, fmt.Errorf("nil store object")
	}

	return &Handler{
		nodeID:     nodeID,
		clusterMap: cmap.New(),
		store:      s,
	}, nil
}
