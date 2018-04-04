package admin

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type handlers struct {
	cfg   *config.Mds
	store Repository
	cMap  *cmap.CMap
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cfg *config.Mds, s Repository) delivery.AdminHandlers {
	log = mlog.GetLogger().WithField("package", "mds/usecase/admin")

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
	// q := fmt.Sprintf(
	// 	`
	// 	SELECT
	//         local_chain_id, parity_volume_id
	// 	FROM
	//         local_chain
	// 	`,
	// )

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
	ctxLogger := log.WithField("method", "handlers.updateClusterMap")

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
