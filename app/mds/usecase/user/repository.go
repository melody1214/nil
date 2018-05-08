package user

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Repository provides access to admin database.
type Repository interface {
	PublishCommand(op, query string) (result sql.Result, err error)
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
	MakeBucket(bucketName, accessKey, region string) (err error)
	FindSecretKey(accessKey string) (secretKey string, err error)
}
