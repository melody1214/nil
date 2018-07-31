package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/domain/model/user"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

type userRepository struct {
	s *Store
}

// NewUserRepository returns a new instance of a mysql user repository.
func NewUserRepository(s *Store) user.Repository {
	return &userRepository{
		s: s,
	}
}

func (r *userRepository) FindByID(id user.ID) (*user.User, error) {
	ctxLogger := mlog.GetMethodLogger(logger, "userRepository.FindByID")

	q := fmt.Sprintf(
		`
		SELECT
			*
		FROM
			user
		WHERE
			user_id='%s'
		`, id.String(),
	)

	u := &user.User{}
	err := r.s.QueryRow(repository.NotTx, q).Scan(u)
	if err == sql.ErrNoRows {
		err = user.ErrNotExist
	} else if err != nil {
		ctxLogger.Error(errors.Wrapf(err, "failed to find user by ID: %s", id.String()))
		err = user.ErrInternal
	}

	return u, err
}

func (r *userRepository) FindByAk(access user.Key) (*user.User, error) {
	ctxLogger := mlog.GetMethodLogger(logger, "userRepository.FindByAk")

	q := fmt.Sprintf(
		`
		SELECT
			*
		FROM
			user
		WHERE
			user_access_key='%s'
		`, access.String(),
	)

	u := &user.User{}
	err := r.s.QueryRow(repository.NotTx, q).Scan(u)
	if err == sql.ErrNoRows {
		err = user.ErrNotExist
	} else if err != nil {
		ctxLogger.Error(errors.Wrapf(err, "failed to find user by access key: %s", access.String()))
		err = user.ErrInternal
	}

	return u, err
}

func (r *userRepository) Save(user *user.User) error {
	if user.ID.String() == "" {
		return r.update(user)
	}
	return r.create(user)
}

func (r *userRepository) update(user *user.User) error {
	q := fmt.Sprintf(
		`
		UPDATE user
		SET user_name='%s', user_secret_key='%s'
		WHERE user_id='%s',
		`, user.Name.String(), user.Secret.String(), user.ID.String(),
	)
	_, err := r.s.PublishCommand("execute", q)
	return err
}

func (r *userRepository) create(user *user.User) error {
	q := fmt.Sprintf(
		`
		INSERT INTO user (user_name, user_access_key, user_secret_key)
		SELECT * FROM (SELECT '%s' AS un, '%s' AS ak, '%s' AS sk) AS tmp
		WHERE NOT EXISTS (
			SELECT user_access_key FROM user WHERE user_access_key = '%s'
		) LIMIT 1;
		`, user.Name.String(), user.Access.String(), user.Secret.String(), user.Access.String(),
	)
	_, err := r.s.PublishCommand("execute", q)
	return err
}
