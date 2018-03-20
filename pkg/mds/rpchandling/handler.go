package rpchandling

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/util/mlog"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

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

	// RegisterVolume adds a new volume from ds into the db and returns a registered volume id.
	RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error
}

// TypeBytes returns rpc type bytes which is used to multiplexing.
func TypeBytes() []byte {
	return []byte{
		0x02, // rpcNil
	}
}

// Handler has exposed methods for rpc server.
type Handler struct {
	store *store.Store
}

// New returns a new rpc handler.
func New(store *store.Store) (NilRPCHandler, error) {
	log = mlog.GetLogger()

	if store == nil {
		return nil, fmt.Errorf("nil server object")
	}

	return &Handler{store: store}, nil
}

// Join joins the mds node into the cluster.
func (h *Handler) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return h.store.Join(req.NodeID, req.RaftAddr)
}

// AddUser adds a new user with the given name.
func (h *Handler) AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
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
	_, err := h.store.PublishCommand("execute", q)
	if err != nil {
		return err
	}

	res.AccessKey = ak.AccessKey()
	res.SecretKey = ak.SecretKey()

	return nil
}

// GetCredential returns matching secret key with the given access key.
func (h *Handler) GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
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
	err := h.store.QueryRow(q).Scan(&res.SecretKey)
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
func (h *Handler) AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	q := fmt.Sprintf(
		`
		INSERT INTO bucket (bucket_name, user_id, region_id)
		SELECT '%s', u.user_id, r.region_id
		FROM user u, region r
		WHERE u.access_key = '%s' and r.region_name = '%s';
		`, req.BucketName, req.AccessKey, req.Region,
	)

	_, err := h.store.PublishCommand("execute", q)
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

// GetClusterMap returns a current local cluster map.
func (h *Handler) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	// TODO: Other routine with the smart cmap.
	// h.updateMembership()
	// cm, err := h.updateClusterMap()
	// if err != nil {
	// 	return err
	// }

	// res.Version = cm.Version

	// for _, n := range cm.Nodes {
	// 	res.Nodes = append(
	// 		res.Nodes,
	// 		nilrpc.ClusterNode{
	// 			ID:   n.ID.Int64(),
	// 			Name: n.Name,
	// 			Addr: n.Addr,
	// 			Type: n.Type.String(),
	// 			Stat: n.Stat.String(),
	// 		},
	// 	)
	// }

	return nil
}

// RegisterVolume receives a new volume information from ds and register it to the database.
func (h *Handler) RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	// If the id field of request is empty, then the ds
	// tries to get an id of volume.
	if req.ID == "" {
		return h.insertNewVolume(req, res)
	}
	return h.updateVolume(req, res)
}

func (h *Handler) updateVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	log.Infof("update a member %v", req)

	q := fmt.Sprintf(
		`
		UPDATE volume
		SET volume_status='%s', size='%d', free='%d', used='%d', speed='%s'
		WHERE volume_id in ('%s')
		`, req.Status, req.Size, req.Free, req.Used, req.Speed, req.ID,
	)

	_, err := h.store.Execute(q)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (h *Handler) insertNewVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	log.Infof("insert a new volume %v", req)

	q := fmt.Sprintf(
		`
		INSERT INTO volume (node_id, volume_status, size, free, used, speed)
		SELECT node_id, '%s', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
		`, req.Status, req.Size, req.Free, req.Used, req.Speed, req.Ds,
	)

	r, err := h.store.Execute(q)
	if err != nil {
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return err
	}
	res.ID = strconv.FormatInt(id, 10)

	return nil
}
