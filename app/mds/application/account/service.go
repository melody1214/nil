package account

import (
	"github.com/chanyoung/nil/app/mds/domain/model/bucket"
	"github.com/chanyoung/nil/app/mds/domain/model/region"
	"github.com/chanyoung/nil/app/mds/domain/model/user"
	"github.com/chanyoung/nil/app/mds/domain/service/raft"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type service struct {
	cfg *config.Mds

	rss raft.SimpleService

	rgr region.Repository
	usr user.Repository
	bkr bucket.Repository
}

// NewService creates a user service with necessary dependencies.
func NewService(cfg *config.Mds, rss raft.SimpleService, rgr region.Repository, usr user.Repository, bkr bucket.Repository) Service {
	logger = mlog.GetPackageLogger("app/mds/usecase/admin")

	sv := &service{
		cfg: cfg,
		rss: rss,
		rgr: rgr,
		usr: usr,
		bkr: bkr,
	}

	return sv
}

// AddUser adds a new user with the given name.
func (s *service) AddUser(req *nilrpc.MACAddUserRequest, res *nilrpc.MACAddUserResponse) error {
	// User is the globally shared metadata.
	// If this node is not a leader but has received a request, it forwards
	// the request to the leader node instead.
	leader, err := s.rss.Leader()
	if err != nil {
		return err
	}
	if !leader {
		leaderEndPoint, err := s.rss.LeaderEndPoint()
		if err != nil {
			return err
		}

		_ = leaderEndPoint
		// 	conn, err := nilrpc.Dial(leaderEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
		// 	if err != nil {
		// 		return err
		// 	}
		// 	defer conn.Close()

		// 	cli := rpc.NewClient(conn)
		// 	defer cli.Close()

		// 	return cli.Call(nilrpc.MdsUserAddUser.String(), req, res)
	}

	u := &user.User{
		Name:   user.Name(req.Name),
		Access: user.GenKey(),
		Secret: user.GenKey(),
	}

	res.AccessKey = u.Access.String()
	res.SecretKey = u.Secret.String()

	return s.usr.Save(u)
}

// MakeBucket creates a bucket with the given name.
func (s *service) MakeBucket(req *nilrpc.MACMakeBucketRequest, res *nilrpc.MACMakeBucketResponse) error {
	// Bucket is the globally shared metadata.
	// If this node is not a leader but has received a request, it forwards
	// the request to the leader node instead.
	leader, err := s.rss.Leader()
	if err != nil {
		return err
	}
	if !leader {
		leaderEndPoint, err := s.rss.LeaderEndPoint()
		if err != nil {
			return err
		}

		_ = leaderEndPoint
		// 	conn, err := nilrpc.Dial(leaderEndpoint, nilrpc.RPCNil, time.Duration(2*time.Second))
		// 	if err != nil {
		// 		return err
		// 	}
		// 	defer conn.Close()

		// 	cli := rpc.NewClient(conn)
		// 	defer cli.Close()

		// 	return cli.Call(nilrpc.MdsUserMakeBucket.String(), req, res)
	}

	u, err := s.usr.FindByAk(user.Key(req.AccessKey))
	if err != nil {
		return err
	}

	r, err := s.rgr.FindByName(region.Name(req.Region))
	if err != nil {
		return err
	}

	err = s.bkr.Save(&bucket.Bucket{
		Name:   bucket.Name(req.BucketName),
		User:   bucket.ID(u.ID),
		Region: bucket.ID(r.ID),
	})

	switch err {
	case nil:
		res.S3ErrCode = s3.ErrNone
	case bucket.ErrDuplicateEntry:
		res.S3ErrCode = s3.ErrBucketAlreadyExists
	default:
		res.S3ErrCode = s3.ErrInternalError
	}

	return err
}

// GetCredential returns matching secret key with the given access key.
func (s *service) GetCredential(req *nilrpc.MACGetCredentialRequest, res *nilrpc.MACGetCredentialResponse) error {
	res.AccessKey = req.AccessKey

	// Find by given access key.
	u, err := s.usr.FindByAk(user.Key(req.AccessKey))
	if err == nil {
		res.Exist = true
		res.SecretKey = u.Secret.String()
	} else if err == user.ErrNotExist {
		res.Exist = false
		err = nil
	}

	return err
}

// Service is the interface that provides user domain's rpc handlers.
type Service interface {
	AddUser(req *nilrpc.MACAddUserRequest, res *nilrpc.MACAddUserResponse) error
	MakeBucket(req *nilrpc.MACMakeBucketRequest, res *nilrpc.MACMakeBucketResponse) error
	GetCredential(req *nilrpc.MACGetCredentialRequest, res *nilrpc.MACGetCredentialResponse) error
}
