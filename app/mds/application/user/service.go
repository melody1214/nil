package user

import (
	"github.com/chanyoung/nil/pkg/nilrpc"
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

	sv := &service{
		cfg:   cfg,
		store: s,
	}

	return sv
}

// AddUser adds a new user with the given name.
func (s *service) AddUser(req *nilrpc.MUSAddUserRequest, res *nilrpc.MUSAddUserResponse) error {
	// // User is the globally shared metadata.
	// // If this node is not a leader but has received a request, it forwards
	// // the request to the leader node instead.
	// leader, err := s.store.AmILeader()
	// if err != nil {
	// 	return err
	// }
	// if !leader {
	// 	leaderEndpoint := s.store.LeaderEndpoint()
	// 	if leaderEndpoint == "" {
	// 		return fmt.Errorf("This node is not leader, and the leader is not exist in global cluster")
	// 	}

	// 	conn, err := nilrpc.Dial(leaderEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer conn.Close()

	// 	cli := rpc.NewClient(conn)
	// 	defer cli.Close()

	// 	return cli.Call(nilrpc.MdsUserAddUser.String(), req, res)
	// }

	// ak := security.NewAPIKey()
	// s.store.AddUser(req.Name, ak)

	// res.AccessKey = ak.AccessKey()
	// res.SecretKey = ak.SecretKey()
	return nil
}

// MakeBucket creates a bucket with the given name.
func (s *service) MakeBucket(req *nilrpc.MUSMakeBucketRequest, res *nilrpc.MUSMakeBucketResponse) error {
	// // Bucket is the globally shared metadata.
	// // If this node is not a leader but has received a request, it forwards
	// // the request to the leader node instead.
	// leader, err := s.store.AmILeader()
	// if err != nil {
	// 	return err
	// }
	// if !leader {
	// 	leaderEndpoint := s.store.LeaderEndpoint()
	// 	if leaderEndpoint == "" {
	// 		return fmt.Errorf("This node is not leader, and the leader is not exist in global cluster")
	// 	}

	// 	conn, err := nilrpc.Dial(leaderEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer conn.Close()

	// 	cli := rpc.NewClient(conn)
	// 	defer cli.Close()

	// 	return cli.Call(nilrpc.MdsUserMakeBucket.String(), req, res)
	// }

	// err = s.store.MakeBucket(req.BucketName, req.AccessKey, req.Region)
	// if err == repository.ErrDuplicateEntry {
	// 	res.S3ErrCode = s3.ErrBucketAlreadyExists
	// } else if err != nil {
	// 	res.S3ErrCode = s3.ErrInternalError
	// }
	return nil
}

// GetCredential returns matching secret key with the given access key.
func (s *service) GetCredential(req *nilrpc.MUSGetCredentialRequest, res *nilrpc.MUSGetCredentialResponse) error {
	// res.AccessKey = req.AccessKey

	// sk, err := s.store.FindSecretKey(req.AccessKey)
	// if err == nil {
	// 	res.Exist = true
	// 	res.SecretKey = sk
	// } else if err == repository.ErrNotExist {
	// 	res.Exist = false
	// } else {
	// 	return err
	// }

	return nil
}

// Service is the interface that provides user domain's rpc handlers.
type Service interface {
	AddUser(req *nilrpc.MUSAddUserRequest, res *nilrpc.MUSAddUserResponse) error
	MakeBucket(req *nilrpc.MUSMakeBucketRequest, res *nilrpc.MUSMakeBucketResponse) error
	GetCredential(req *nilrpc.MUSGetCredentialRequest, res *nilrpc.MUSGetCredentialResponse) error
}
