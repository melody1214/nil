package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/domain/model/user"
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
	return nil, nil
}

func (r *userRepository) FindByAk(access user.Key) (*user.User, error) {
	return nil, nil
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

// func (s *userStore) MakeBucket(bucketName, accessKey, region string) (err error) {
// 	q := fmt.Sprintf(
// 		`
// 		INSERT INTO bucket (bk_name, bk_user, bk_region)
// 		SELECT '%s', u.user_id, r.rg_id
// 		FROM user u, region r
// 		WHERE u.user_access_key = '%s' and r.rg_name = '%s';
// 		`, bucketName, accessKey, region,
// 	)

// 	_, err = s.PublishCommand("execute", q)
// 	// No error occurred while adding the bucket.
// 	if err == nil {
// 		return nil
// 	}
// 	// Error occurred.
// 	mysqlError, ok := err.(*mysql.MySQLError)
// 	if !ok {
// 		// Not mysql error occurred, return itself.
// 		return err
// 	}

// 	// Mysql error occurred. Classify it and sending the corresponding s3 error code.
// 	switch mysqlError.Number {
// 	case 1062:
// 		return repository.ErrDuplicateEntry
// 	default:
// 		return err
// 	}
// }

// func (s *userStore) FindSecretKey(accessKey string) (secretKey string, err error) {
// 	q := fmt.Sprintf(
// 		`
// 		SELECT
// 			user_secret_key
// 		FROM
// 			user
// 		WHERE
// 			user_access_key = '%s'
// 		`, accessKey,
// 	)

// 	err = s.QueryRow(repository.NotTx, q).Scan(&secretKey)
// 	if err == sql.ErrNoRows {
// 		err = repository.ErrNotExist
// 	}
// 	return
// }
