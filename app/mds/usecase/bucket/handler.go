package bucket

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/delivery"
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
func NewHandlers(s Repository) delivery.BucketHandlers {
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
		INSERT INTO bucket (bucket_name, user_id, region_id)
		SELECT '%s', u.user_id, r.region_id
		FROM user u, region r
		WHERE u.access_key = '%s' and r.region_name = '%s';
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
