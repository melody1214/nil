package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/domain/model/bucket"
	"github.com/go-sql-driver/mysql"
)

type bucketRepository struct {
	s *Store
}

// NewBucketRepository returns a new instance of a mysql bucket repository.
func NewBucketRepository(s *Store) bucket.Repository {
	return &bucketRepository{
		s: s,
	}
}

func (r *bucketRepository) Save(b *bucket.Bucket) error {
	if b.ID.String() == "" {
		return r.update(b)
	}
	return r.create(b)
}

func (r *bucketRepository) update(b *bucket.Bucket) error {
	q := fmt.Sprintf(
		`
		UPDATE bucket
		SET bk_name='%s', bk_user='%s', bk_region='%s',
		WHERE bk_id='%s',
		`, b.Name.String(), b.User.String(), b.Region.String(), b.ID.String(),
	)
	_, err := r.s.PublishCommand("execute", q)
	return err
}

func (r *bucketRepository) create(b *bucket.Bucket) error {
	q := fmt.Sprintf(
		`
		INSERT INTO bucket (bk_name, bk_user, bk_region)
		VALUES ('%s', '%s', '%s')
		`, b.Name.String(), b.User.String(), b.Region.String(),
	)

	_, err := r.s.PublishCommand("execute", q)
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
		return bucket.ErrDuplicateEntry
	default:
		return err
	}
}
