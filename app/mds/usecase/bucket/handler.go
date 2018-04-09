package bucket

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/go-sql-driver/mysql"
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

// TODO: CQRS
// AddBucket creates a bucket with the given name.
func (h *handlers) AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	q := fmt.Sprintf(
		`
		INSERT INTO bucket (bk_name, bk_user, bk_region)
		SELECT '%s', u.user_id, r.rg_id
		FROM user u, region r
		WHERE u.user_access_key = '%s' and r.rg_name = '%s';
		`, req.BucketName, req.AccessKey, req.Region,
	)

	_, err := h.store.PublishCommand("execute", q)
	// No error occurred while adding the bucket.
	if err == nil {
		res.S3ErrCode = s3.ErrNone
		return nil
	}
	// Error occurred.
	mysqlError, ok := err.(*mysql.MySQLError)
	if !ok {
		// Not mysql error occurred, return itself.
		return err
	}

	// Mysql error occurred. Classify it and sending the corresponding s3 error code.
	switch mysqlError.Number {
	case 1062:
		res.S3ErrCode = s3.ErrBucketAlreadyExists
	default:
		res.S3ErrCode = s3.ErrInternalError
	}
	return nil
}

// Handlers is the interface that provides bucket domain's rpc handlers.
type Handlers interface {
	AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error
}
