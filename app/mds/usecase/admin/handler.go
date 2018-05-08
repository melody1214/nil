package admin

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	cfg   *config.Mds
	store Repository
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/admin")

	return &handlers{
		cfg:   cfg,
		store: s,
	}
}

// TODO: CQRS

// Join joins the mds node into the cluster.
func (h *handlers) Join(req *nilrpc.MADJoinRequest, res *nilrpc.MADJoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}
	return h.store.Join(req.NodeID, req.RaftAddr)
}

// AddUser adds a new user with the given name.
func (h *handlers) AddUser(req *nilrpc.MADAddUserRequest, res *nilrpc.MADAddUserResponse) error {
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
	_, err := h.store.PublishCommand("execute", q)
	if err != nil {
		return err
	}

	res.AccessKey = ak.AccessKey()
	res.SecretKey = ak.SecretKey()

	return nil
}

// RegisterVolume receives a new volume information from ds and register it to the database.
func (h *handlers) RegisterVolume(req *nilrpc.MADRegisterVolumeRequest, res *nilrpc.MADRegisterVolumeResponse) error {
	// If the id field of request is empty, then the ds
	// tries to get an id of volume.
	if req.ID == "" {
		return h.insertNewVolume(req, res)
	}
	return h.updateVolume(req, res)
}

func (h *handlers) updateVolume(req *nilrpc.MADRegisterVolumeRequest, res *nilrpc.MADRegisterVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateVolume")

	q := fmt.Sprintf(
		`
		UPDATE volume
		SET vl_status='%s', vl_size='%d', vl_free='%d', vl_used='%d', vl_max_encoding_group='%d', vl_speed='%s' 
		WHERE vl_id in ('%s')
		`, req.Status, req.Size, req.Free, req.Used, calcMaxChain(req.Size), req.Speed, req.ID,
	)

	_, err := h.store.Execute(repository.NotTx, q)
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	return h.updateClusterMap()
}

func (h *handlers) insertNewVolume(req *nilrpc.MADRegisterVolumeRequest, res *nilrpc.MADRegisterVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.insertNewVolume")

	q := fmt.Sprintf(
		`
		INSERT INTO volume (vl_node, vl_status, vl_size, vl_free, vl_used, vl_encoding_group, vl_max_encoding_group, vl_speed)
		SELECT node_id, '%s', '%d', '%d', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
		`, req.Status, req.Size, req.Free, req.Used, 0, calcMaxChain(req.Size), req.Speed, req.Ds,
	)

	r, err := h.store.Execute(repository.NotTx, q)
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		ctxLogger.Error(err)
		return err
	}
	res.ID = strconv.FormatInt(id, 10)

	return h.updateClusterMap()
}

func (h *handlers) updateClusterMap() error {
	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MCLUpdateClusterMapRequest{}
	res := &nilrpc.MCLUpdateClusterMapResponse{}

	cli := rpc.NewClient(conn)
	defer cli.Close()

	return cli.Call(nilrpc.MdsClustermapUpdateClusterMap.String(), req, res)
}

func calcMaxChain(volumeSize uint64) int {
	if volumeSize <= 0 {
		return 0
	}

	// Test, chain per 10MB,
	return int(volumeSize / 10)
}

// Handlers is the interface that provides admin domain's rpc handlers.
type Handlers interface {
	Join(req *nilrpc.MADJoinRequest, res *nilrpc.MADJoinResponse) error
	AddUser(req *nilrpc.MADAddUserRequest, res *nilrpc.MADAddUserResponse) error
	RegisterVolume(req *nilrpc.MADRegisterVolumeRequest, res *nilrpc.MADRegisterVolumeResponse) error
}
