package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/bucket"
	"github.com/go-sql-driver/mysql"
)

type bucketStore struct {
	*Store
}

// NewBucketRepository returns a new instance of a mysql bucket repository.
func NewBucketRepository(s *Store) bucket.Repository {
	return &bucketStore{
		Store: s,
	}
}

func (s *bucketStore) MakeBucket(bucketName, accessKey, region string) (err error) {
	q := fmt.Sprintf(
		`
		INSERT INTO bucket (bk_name, bk_user, bk_region)
		SELECT '%s', u.user_id, r.rg_id
		FROM user u, region r
		WHERE u.user_access_key = '%s' and r.rg_name = '%s';
		`, bucketName, accessKey, region,
	)

	_, err = s.PublishCommand("execute", q)
	// No error occurred while adding the bucket.
	if err == nil {
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
		return repository.ErrDuplicateEntry
	default:
		return err
	}
}
