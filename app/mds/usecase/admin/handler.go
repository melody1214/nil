package admin

import (
	"fmt"
	"strconv"

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
func NewHandlers(cfg *config.Mds, s Repository) AdminHandlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/admin")

	return &handlers{
		cfg:   cfg,
		store: s,
	}
}

// TODO: CQRS

// Join joins the mds node into the cluster.
func (h *handlers) Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error {
	if req.RaftAddr == "" || req.NodeID == "" {
		return fmt.Errorf("not enough arguments: %+v", req)
	}
	return h.store.Join(req.NodeID, req.RaftAddr)
}

// GetLocalChain : Test code. Will be removed soon.
func (h *handlers) GetLocalChain(req *nilrpc.GetLocalChainRequest, res *nilrpc.GetLocalChainResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
	        eg_id, eg_parity_volume
		FROM
	        encoding_group
	    ORDER BY rand() limit 1;
		`,
	)

	row := h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	err := row.Scan(&res.LocalChainID, &res.ParityVolumeID)
	if err != nil {
		return err
	}

	q = fmt.Sprintf(
		`
		SELECT
			vl_node
			FROM
			volume
 		WHERE
			vl_id = '%d'
		`, res.ParityVolumeID,
	)

	row = h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	return row.Scan(&res.ParityNodeID)
}

func (h *handlers) GetAllChain(req *nilrpc.GetAllChainRequest, res *nilrpc.GetAllChainResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
			eg_id, eg_first_volume, eg_second_volume, eg_third_volume, eg_parity_volume
		FROM
			encoding_group
		`,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		c := nilrpc.Chain{}

		if err := rows.Scan(&c.ID, &c.FirstVolumeID, &c.SecondVolumeID, &c.ThirdVolumeID, &c.ParityVolumeID); err != nil {
			return err
		}

		res.Chains = append(res.Chains, c)
	}

	return nil
}

func (h *handlers) GetAllVolume(req *nilrpc.GetAllVolumeRequest, res *nilrpc.GetAllVolumeResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
			vl_id, vl_node
		FROM
			volume
		`,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		v := nilrpc.Volume{}

		if err := rows.Scan(&v.ID, &v.NodeID); err != nil {
			return err
		}

		res.Volumes = append(res.Volumes, v)
	}

	return nil
}

// AddUser adds a new user with the given name.
func (h *handlers) AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
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
func (h *handlers) RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	// If the id field of request is empty, then the ds
	// tries to get an id of volume.
	if req.ID == "" {
		return h.insertNewVolume(req, res)
	}
	return h.updateVolume(req, res)
}

func (h *handlers) updateVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateVolume")

	q := fmt.Sprintf(
		`
		UPDATE volume
		SET vl_status='%s', vl_size='%d', vl_free='%d', vl_used='%d', vl_max_encoding_group='%d', vl_speed='%s' 
		WHERE vl_id in ('%s')
		`, req.Status, req.Size, req.Free, req.Used, calcMaxChain(req.Size), req.Speed, req.ID,
	)

	_, err := h.store.Execute(q)
	if err != nil {
		ctxLogger.Error(err)
	}
	return err
}

func (h *handlers) insertNewVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.insertNewVolume")

	q := fmt.Sprintf(
		`
		INSERT INTO volume (vl_node, vl_status, vl_size, vl_free, vl_used, vl_encoding_group, vl_max_encoding_group, vl_speed)
		SELECT node_id, '%s', '%d', '%d', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
		`, req.Status, req.Size, req.Free, req.Used, 0, calcMaxChain(req.Size), req.Speed, req.Ds,
	)

	r, err := h.store.Execute(q)
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

	return nil
}

func calcMaxChain(volumeSize uint64) int {
	if volumeSize <= 0 {
		return 0
	}

	// Test, chain per 10MB,
	return int(volumeSize / 10)
}

// AdminHandlers is the interface that provides admin domain's rpc handlers.
type AdminHandlers interface {
	Join(req *nilrpc.JoinRequest, res *nilrpc.JoinResponse) error
	AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error
	GetLocalChain(req *nilrpc.GetLocalChainRequest, res *nilrpc.GetLocalChainResponse) error
	GetAllChain(req *nilrpc.GetAllChainRequest, res *nilrpc.GetAllChainResponse) error
	GetAllVolume(req *nilrpc.GetAllVolumeRequest, res *nilrpc.GetAllVolumeResponse) error
	RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error
}
