package user

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg   *config.Mds
	store Repository
}

// NewService creates a user service with necessary dependencies.
func NewService(cfg *config.Mds, s Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/admin")

	return &service{
		cfg:   cfg,
		store: s,
	}
}

// TODO: CQRS

// AddUser adds a new user with the given name.
func (s *service) AddUser(req *nilrpc.MUSAddUserRequest, res *nilrpc.MUSAddUserResponse) error {
	ak := security.NewAPIKey()

	q := fmt.Sprintf(
		`
		INSERT INTO user (user_name, user_access_key, user_secret_key)
		SELECT * FROM (SELECT '%s' AS un, '%s' AS ak, '%s' AS sk) AS tmp
		WHERE NOT EXISTS (
			SELECT user_name FROM user WHERE user_name = '%s'
		) LIMIT 1;
		`, req.Name, ak.AccessKey(), ak.SecretKey(), req.Name,
	)
	_, err := s.store.PublishCommand("execute", q)
	if err != nil {
		return err
	}

	res.AccessKey = ak.AccessKey()
	res.SecretKey = ak.SecretKey()

	return nil
}

// MakeBucket creates a bucket with the given name.
func (s *service) MakeBucket(req *nilrpc.MUSMakeBucketRequest, res *nilrpc.MUSMakeBucketResponse) error {
	err := s.store.MakeBucket(req.BucketName, req.AccessKey, req.Region)
	if err == repository.ErrDuplicateEntry {
		res.S3ErrCode = s3.ErrBucketAlreadyExists
	} else if err != nil {
		res.S3ErrCode = s3.ErrInternalError
	}

	return nil
}

// GetCredential returns matching secret key with the given access key.
func (s *service) GetCredential(req *nilrpc.MUSGetCredentialRequest, res *nilrpc.MUSGetCredentialResponse) error {
	res.AccessKey = req.AccessKey

	sk, err := s.store.FindSecretKey(req.AccessKey)
	if err == nil {
		res.Exist = true
		res.SecretKey = sk
	} else if err == repository.ErrNotExist {
		res.Exist = false
	} else {
		return err
	}

	return nil
}

// Service is the interface that provides user domain's rpc handlers.
type Service interface {
	AddUser(req *nilrpc.MUSAddUserRequest, res *nilrpc.MUSAddUserResponse) error
	MakeBucket(req *nilrpc.MUSMakeBucketRequest, res *nilrpc.MUSMakeBucketResponse) error
	GetCredential(req *nilrpc.MUSGetCredentialRequest, res *nilrpc.MUSGetCredentialResponse) error
}
