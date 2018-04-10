package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/auth"
)

type authStore struct {
	*Store
}

// NewAuthRepository returns a new instance of a mysql auth repository.
func NewAuthRepository(s *Store) auth.Repository {
	return &authStore{
		Store: s,
	}
}

func (s *authStore) FindSecretKey(accessKey string) (secretKey string, err error) {
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
