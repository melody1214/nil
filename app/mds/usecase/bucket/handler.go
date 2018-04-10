package bucket

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store Repository
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/bucket")

	return &handlers{
		store: s,
	}
}

// MakeBucket creates a bucket with the given name.
func (h *handlers) MakeBucket(req *nilrpc.MBUMakeBucketRequest, res *nilrpc.MBUMakeBucketResponse) error {
	err := h.store.MakeBucket(req.BucketName, req.AccessKey, req.Region)
	if err == repository.ErrDuplicateEntry {
		res.S3ErrCode = s3.ErrBucketAlreadyExists
	} else if err != nil {
		res.S3ErrCode = s3.ErrInternalError
	}

	return nil
}

// Handlers is the interface that provides bucket domain's rpc handlers.
type Handlers interface {
	MakeBucket(req *nilrpc.MBUMakeBucketRequest, res *nilrpc.MBUMakeBucketResponse) error
}
