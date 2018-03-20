package mds

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilmux"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/go-sql-driver/mysql"
)

// Handler has exposed methods for rpc server.
type Handler struct {
	m *Mds
}

func newNilRPCHandler(m *Mds) (NilRPCHandler, error) {
	if m == nil {
		return nil, fmt.Errorf("nil server object")
	}

	return &Handler{m: m}, nil
}

// Join is an exposed method of swim rpc service.
// It simply wraps the server's handleJoin method.
func (h *Handler) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	return h.m.handleJoin(req, res)
}

// AddUser is an exposed method of swim rpc service.
// It simply wraps the server's handleAddUser method.
func (h *Handler) AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
	return h.m.handleAddUser(req, res)
}

// GetCredential is an exposed method of swim rpc service.
// It simply wraps the server's handleGetCredential method.
func (h *Handler) GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
	return h.m.handleGetCredential(req, res)
}

// AddBucket is an exposed method of swim rpc service.
// It simply wraps the server's handleAddBucket method.
func (h *Handler) AddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	return h.m.handleAddBucket(req, res)
}

// GetClusterMap is an exposed method of swim rpc service.
// It simply wraps the server's handleGetClusterMap method.
func (h *Handler) GetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	return h.m.handleGetClusterMap(req, res)
}

// RegisterVolume receives a new volume information from ds and register it to the database.
func (h *Handler) RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	return h.m.handleRegisterVolume(req, res)
}

func (m *Mds) serveNilRPC(l *nilmux.Layer) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go m.nilRPCSrv.ServeConn(conn)
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

	// RegisterVolume adds a new volume from ds into the db and returns a registered volume id.
	RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error
}

// handleJoin joins the mds node into the cluster.
func (m *Mds) handleJoin(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}

	return m.store.Join(req.NodeID, req.RaftAddr)
}

// handleAddUser adds a new user with the given name.
func (m *Mds) handleAddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
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
	_, err := m.store.PublishCommand("execute", q)
	if err != nil {
		return err
	}

	res.AccessKey = ak.AccessKey()
	res.SecretKey = ak.SecretKey()

	return nil
}

// handleGetCredential returns matching secret key with the given access key.
func (m *Mds) handleGetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
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
	err := m.store.QueryRow(q).Scan(&res.SecretKey)
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
func (m *Mds) handleAddBucket(req *nilrpc.AddBucketRequest, res *nilrpc.AddBucketResponse) error {
	log.Infof("%+v", req)

	q := fmt.Sprintf(
		`
		INSERT INTO bucket (bucket_name, user_id, region_id)
		SELECT '%s', u.user_id, r.region_id
		FROM user u, region r
		WHERE u.access_key = '%s' and r.region_name = '%s';
		`, req.BucketName, req.AccessKey, m.cfg.Raft.LocalClusterRegion,
	)

	log.Infof("%+v", q)

	_, err := m.store.PublishCommand("execute", q)
	log.Infof("%+v", err)
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

func (m *Mds) handleRegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	// If the id field of request is empty, then the ds
	// tries to get an id of volume.
	if req.ID == "" {
		return m.insertNewVolume(req, res)
	}
	return m.updateVolume(req, res)
}

func (m *Mds) updateVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	log.Infof("update a member %v", req)

	q := fmt.Sprintf(
		`
		UPDATE volume
		SET volume_status='%s', size='%d', free='%d', used='%d', speed='%s'
		WHERE volume_id in ('%s')
		`, req.Status, req.Size, req.Free, req.Used, req.Speed, req.ID,
	)

	_, err := m.store.Execute(q)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (m *Mds) insertNewVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	log.Infof("insert a new volume %v", req)

	q := fmt.Sprintf(
		`
		INSERT INTO volume (node_id, volume_status, size, free, used, speed)
		SELECT node_id, '%s', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
		`, req.Status, req.Size, req.Free, req.Used, req.Speed, req.Ds,
	)

	r, err := m.store.Execute(q)
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

// handleGetClusterMap returns a current local cluster map.
func (m *Mds) handleGetClusterMap(req *nilrpc.GetClusterMapRequest, res *nilrpc.GetClusterMapResponse) error {
	cm, err := cmap.GetLatest()
	if err != nil {
		return err
	}

	res.Version = cm.Version
	for _, n := range cm.Nodes {
		res.Nodes = append(
			res.Nodes,
			nilrpc.ClusterNode{
				ID:   n.ID.Int64(),
				Name: n.Name,
				Addr: n.Addr,
				Type: n.Type.String(),
				Stat: n.Stat.String(),
			},
		)
	}

	return nil
}
