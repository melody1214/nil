package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/cluster"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/pkg/errors"
)

type clusterStore struct {
	*Store
}

// NewClusterRepository returns a new instance of a mysql cluster map repository.
func NewClusterRepository(s *Store) cluster.Repository {
	return &clusterStore{
		Store: s,
	}
}

func (s *clusterStore) FindAllNodes(txid repository.TxID) (nodes []cmap.Node, err error) {
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

func (s *clusterStore) FindAllVolumes(txid repository.TxID) (vols []cmap.Volume, err error) {
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

func (s *clusterStore) FindAllEncGrps(txid repository.TxID) (egs []cmap.EncodingGroup, err error) {
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

func (s *clusterStore) FindAllEncGrpVols(txid repository.TxID, id cmap.ID) (vols []cmap.ID, err error) {
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

func (s *clusterStore) GetNewClusterMapVer(txid repository.TxID) (cmap.Version, error) {
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

func (s *clusterStore) LocalJoin(node cmap.Node) error {
	q := fmt.Sprintf(
		`
		INSERT INTO node (node_name, node_type, node_status, node_address)
		VALUES ('%s', '%s', '%s', '%s')
		`, node.Name, node.Type, node.Stat, node.Addr,
	)

	_, err := s.Execute(repository.NotTx, q)
	return err
}

func (s *clusterStore) GlobalJoin(raftAddr, nodeID string) error {
	return s.Store.Join(nodeID, raftAddr)
}

func (s *clusterStore) InsertJob(txid repository.TxID, job *cluster.Job) error {
	if job.Type == cluster.Batch {
		// MergeJob.
		if err := s.mergeJob(txid, &job.Event); err != nil {
			return errors.Wrap(err, "failed to merge old jobs")
		}
	}

	var q string
	if job.Event.AffectedEG == cluster.NoAffectedEG {
		q = fmt.Sprintf(
			`
		INSERT INTO cluster_job (clj_type, clj_state, clj_event_type, clj_event_time)
		VALUES ('%d', '%d', '%d', '%s')
		`, job.Type, job.State, job.Event.Type, job.Event.TimeStamp,
		)
	} else {
		q = fmt.Sprintf(
			`
		INSERT INTO cluster_job (clj_type, clj_state, clj_event_type, clj_event_affected, clj_event_time)
		VALUES ('%d', '%d', '%d', '%s', '%s')
		`, job.Type, job.State, job.Event.Type, job.Event.AffectedEG, job.Event.TimeStamp,
		)
	}

	r, err := s.Execute(txid, q)
	if err != nil {
		return errors.Wrap(err, "failed to insert job")
	}
	id, err := r.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to insert job")
	}
	job.ID = cluster.ID(id)
	return nil
}

func (s *clusterStore) mergeJob(txid repository.TxID, event *cluster.Event) error {
	affectedEG := event.AffectedEG.String()
	if event.AffectedEG.Int64() < 0 {
		affectedEG = "NULL"
	}

	q := fmt.Sprintf(
		`
		UPDATE cluster_job
		SET clj_state=%d, clj_finished='%s'
		WHERE clj_event_type=%d AND clj_event_affected=%s AND clj_state=%d
		`, cluster.Merged, cluster.TimeNow(), event.Type, affectedEG, cluster.Ready,
	)

	_, err := s.Execute(txid, q)
	return err
}
