package admin

import (
	"fmt"
	"strconv"

	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/pkg/cmap"
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
	cMap  *cmap.CMap
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, s Repository) delivery.AdminHandlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/admin")

	return &handlers{
		cfg:   cfg,
		store: s,
		cMap:  cmap.New(),
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
	        local_chain_id, parity_volume_id
		FROM
	        local_chain
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
			node_id
			FROM
			volume
 		WHERE
			volume_id = '%d'
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
			local_chain_id, first_volume_id, second_volume_id, third_volume_id, parity_volume_id
		FROM
			local_chain
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
			volume_id, node_id
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

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	h.cMap = m
}

// AddUser adds a new user with the given name.
func (h *handlers) AddUser(req *nilrpc.AddUserRequest, res *nilrpc.AddUserResponse) error {
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
		SET volume_status='%s', size='%d', free='%d', used='%d', max_chain='%d', speed='%s' 
		WHERE volume_id in ('%s')
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
		INSERT INTO volume (node_id, volume_status, size, free, used, chain, max_chain, speed)
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
