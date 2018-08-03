package object

import (
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
}

// NewService creates a object service with necessary dependencies.
func NewService() Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/object")

	return &service{}
}

func (s *service) Put(req *nilrpc.MOBObjectPutRequest, res *nilrpc.MOBObjectPutResponse) error {
	return nil
}

func (s *service) Get(req *nilrpc.MOBObjectGetRequest, res *nilrpc.MOBObjectGetResponse) error {
	return nil
}

func (s *service) PutChunk(req *nilrpc.MOBPutChunkRequest, res *nilrpc.MOBPutChunkResponse) error {
	return nil
}

func (s *service) GetChunk(req *nilrpc.MOBGetChunkRequest, res *nilrpc.MOBGetChunkResponse) error {
	return nil
}

// Service is the interface that provides object domain's rpc handlers.
type Service interface {
	Put(req *nilrpc.MOBObjectPutRequest, res *nilrpc.MOBObjectPutResponse) error
	Get(req *nilrpc.MOBObjectGetRequest, res *nilrpc.MOBObjectGetResponse) error
	PutChunk(req *nilrpc.MOBPutChunkRequest, res *nilrpc.MOBPutChunkResponse) error
	GetChunk(req *nilrpc.MOBGetChunkRequest, res *nilrpc.MOBGetChunkResponse) error
}
