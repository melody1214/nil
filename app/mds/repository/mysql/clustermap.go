package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/clustermap"
	"github.com/chanyoung/nil/pkg/cmap"
)

type clustermapStore struct {
	*Store
}

// NewClusterMapRepository returns a new instance of a mysql cluster map repository.
func NewClusterMapRepository(s *Store) clustermap.Repository {
	return &clustermapStore{
		Store: s,
	}
}

func (s *clustermapStore) FindAllNodes(txid repository.TxID) (nodes []cmap.Node, err error) {
	q := fmt.Sprintf(
		`
		SELECT
			node_id,
			node_name,
            node_type,
			node_status,
			node_address
		FROM
			node
		`,
	)

	var rows *sql.Rows
	rows, err = s.Query(txid, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes = make([]cmap.Node, 0)
	for rows.Next() {
		n := cmap.Node{}

		if err = rows.Scan(&n.ID, &n.Name, &n.Type, &n.Stat, &n.Addr); err != nil {
			return nil, err
		}

		nodes = append(nodes, n)
	}

	return
}

func (s *clustermapStore) FindAllVolumes(txid repository.TxID) (vols []cmap.Volume, err error) {
	q := fmt.Sprintf(
		`
		SELECT
			vl_id,
			vl_status,
			vl_node,
			vl_size,
			vl_speed
		FROM
			volume
		`,
	)

	var rows *sql.Rows
	rows, err = s.Query(txid, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vols = make([]cmap.Volume, 0)
	for rows.Next() {
		v := cmap.Volume{}

		if err = rows.Scan(&v.ID, &v.Stat, &v.Node, &v.Size, &v.Speed); err != nil {
			return nil, err
		}

		vols = append(vols, v)
	}

	return
}

func (s *clustermapStore) FindAllEncGrps(txid repository.TxID) (egs []cmap.EncodingGroup, err error) {
	q := fmt.Sprintf(
		`
		SELECT
			eg_id,
			eg_status
		FROM
			encoding_group
		`,
	)

	var rows *sql.Rows
	rows, err = s.Query(txid, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	egs = make([]cmap.EncodingGroup, 0)
	for rows.Next() {
		eg := cmap.EncodingGroup{}

		if err = rows.Scan(&eg.ID, &eg.Stat); err != nil {
			return nil, err
		}

		if eg.Vols, err = s.FindAllEncGrpVols(txid, eg.ID); err != nil {
			return nil, err
		}

		egs = append(egs, eg)
	}

	return
}

func (s *clustermapStore) FindAllEncGrpVols(txid repository.TxID, id cmap.ID) (vols []cmap.ID, err error) {
	q := fmt.Sprintf(
		`
		SELECT
			egv_volume
		FROM
			encoding_group_volume
		WHERE
			egv_encoding_group = '%s'
		ORDER BY 
			egv_role DESC
		`, id.String(),
	)

	var rows *sql.Rows
	rows, err = s.Query(txid, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vols = make([]cmap.ID, 0)
	for rows.Next() {
		var volID cmap.ID

		if err = rows.Scan(&volID); err != nil {
			return nil, err
		}

		vols = append(vols, volID)
	}

	return
}

func (s *clustermapStore) GetNewClusterMapVer(txid repository.TxID) (cmap.Version, error) {
	q := fmt.Sprintf(
		`
		INSERT INTO cmap (cmap_id)
		VALUES (NULL)
		`,
	)

	res, err := s.Execute(txid, q)
	if err != nil {
		return -1, err
	}

	ver, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return cmap.Version(ver), nil
}

func (s *clustermapStore) JoinNewNode(node cmap.Node) error {
	q := fmt.Sprintf(
		`
		INSERT INTO node (node_name, node_type, node_status, node_address)
		VALUES ('%s', '%s', '%s', '%s')
		`, node.Name.String(), node.Type.String(), node.Stat.String(), node.Addr.String(),
	)

	_, err := s.Execute(repository.NotTx, q)
	return err
}
