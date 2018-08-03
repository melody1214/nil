package object

import (
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
}

// NewHandlers creates a object handlers with necessary dependencies.
func NewHandlers() Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/object")

	return &handlers{}
}

func (h *handlers) Put(req *nilrpc.MOBObjectPutRequest, res *nilrpc.MOBObjectPutResponse) error {
	return nil
}

func (h *handlers) Get(req *nilrpc.MOBObjectGetRequest, res *nilrpc.MOBObjectGetResponse) error {
	return nil
}

func (h *handlers) PutChunk(req *nilrpc.MOBPutChunkRequest, res *nilrpc.MOBPutChunkResponse) error {
	return nil
}

func (h *handlers) GetChunk(req *nilrpc.MOBGetChunkRequest, res *nilrpc.MOBGetChunkResponse) error {
	return nil
}

// Handlers is the interface that provides object domain's rpc handlers.
type Handlers interface {
	Put(req *nilrpc.MOBObjectPutRequest, res *nilrpc.MOBObjectPutResponse) error
	Get(req *nilrpc.MOBObjectGetRequest, res *nilrpc.MOBObjectGetResponse) error
	PutChunk(req *nilrpc.MOBPutChunkRequest, res *nilrpc.MOBPutChunkResponse) error
	GetChunk(req *nilrpc.MOBGetChunkRequest, res *nilrpc.MOBGetChunkResponse) error
}
