package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/user"
	"github.com/go-sql-driver/mysql"
)

type userStore struct {
	*Store
}

// NewUserRepository returns a new instance of a mysql user repository.
func NewUserRepository(s *Store) user.Repository {
	return &userStore{
		Store: s,
	}
}

func (s *userStore) MakeBucket(bucketName, accessKey, region string) (err error) {
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

func (s *userStore) FindSecretKey(accessKey string) (secretKey string, err error) {
	q := fmt.Sprintf(
		`
		SELECT
			user_secret_key
		FROM
			user
		WHERE
			user_access_key = '%s'
		`, accessKey,
	)

	err = s.QueryRow(repository.NotTx, q).Scan(&secretKey)
	if err == sql.ErrNoRows {
		err = repository.ErrNotExist
	}
	return
}
