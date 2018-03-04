package server

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/go-sql-driver/mysql"
)

// Handler has exposed methods for rpc server.
type Handler struct {
	s *Server
}

func newNilRPCHandler(s *Server) (NilRPCHandler, error) {
	if s == nil {
		return nil, fmt.Errorf("nil server object")
	}

	return &Handler{s: s}, nil
}

// Join is an exposed method of swim rpc service.
// It simply wraps the server's handleJoin method.
func (h *Handler) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	return h.s.handleJoin(req, res)
}

// AddUser is an exposed method of swim rpc service.
// It simply wraps the server's handleAddUser method.
func (h *Handler) AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
	return h.s.handleAddUser(req, res)
}

// GetCredential is an exposed method of swim rpc service.
// It simply wraps the server's handleGetCredential method.
func (h *Handler) GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
	return h.s.handleGetCredential(req, res)
}

// AddBucket is an exposed method of swim rpc service.
// It simply wraps the server's handleAddBucket method.
func (h *Handler) AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	return h.s.handleAddBucket(req, res)
}

// GetClusterMap is an exposed method of swim rpc service.
// It simply wraps the server's handleGetClusterMap method.
func (h *Handler) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	return h.s.handleGetClusterMap(req, res)
}

func (s *Server) serveNilRPC(l *nilmux.Layer) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go s.nilRPCSrv.ServeConn(conn)
	}
}

// NilRPCHandler is the interface of mds rpc commands.
type NilRPCHandler interface {
	// Join joins the mds node into the cluster.
	Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error

	// AddUser adds a new user with the given name.
	AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error
	GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error
	AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error

	// GetClusterMap returns a current local cluster map.
	GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error
}

// handleJoin joins the mds node into the cluster.
func (s *Server) handleJoin(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return s.store.Join(req.NodeID, req.RaftAddr)
}

// handleAddUser adds a new user with the given name.
func (s *Server) handleAddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
	ak := security.NewAPIKey()

	q := fmt.Sprintf(
		`
		INSERT INTO user (user_name, access_key, secret_key)
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

// handleGetCredential returns matching secret key with the given access key.
func (s *Server) handleGetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
			secret_key
		FROM
			user
		WHERE
			access_key = '%s'
		`, req.AccessKey,
	)

	res.AccessKey = req.AccessKey
	err := s.store.QueryRow(q).Scan(&res.SecretKey)
	if err == nil {
		res.Exist = true
	} else if err == sql.ErrNoRows {
		res.Exist = false
	} else {
		return err
	}

	return nil
}

// handleAddBucket creates a bucket with the given name.
func (s *Server) handleAddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	q := fmt.Sprintf(
		`
		INSERT INTO bucket (bucket_name, user_id, region_id)
		SELECT '%s', u.user_id, r.region_id
		FROM user u, region r
		WHERE u.access_key = '%s' and r.region_name = '%s';
		`, req.BucketName, req.AccessKey, s.cfg.Raft.LocalClusterRegion,
	)

	_, err := s.store.PublishCommand("execute", q)
	// No error occurred while adding the bucket.
	if err == nil {
		res.S3ErrCode = s3.ErrNone
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
		res.S3ErrCode = s3.ErrBucketAlreadyExists
	default:
		res.S3ErrCode = s3.ErrInternalError
	}
	return nil
}

// handleGetClusterMap returns a current local cluster map.
func (s *Server) handleGetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	res.Members = s.swimSrv.GetMap()
	return nil
}
