package server

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/security"
)

func (s *Server) newNilRPCHandler() {
	s.NilRPCHandler = s
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

// Join joins the mds node into the cluster.
func (s *Server) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return s.store.Join(req.NodeID, req.RaftAddr)
}

// AddUser adds a new user with the given name.
func (s *Server) AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
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

// GetCredential returns matching secret key with the given access key.
func (s *Server) GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
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

// AddBucket creates a bucket with the given name.
func (s *Server) AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	return nil
}

// GetClusterMap returns a current local cluster map.
func (s *Server) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	res.Members = s.swimSrv.GetMap()
	return nil
}
