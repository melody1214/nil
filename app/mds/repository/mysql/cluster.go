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

	for i, n := range nodes {
		q = fmt.Sprintf(
			`
			SELECT
				vl_id
			FROM
				volume
			WHERE
				vl_node='%d'
			`, n.ID,
		)

		var vID cmap.ID
		rows, err = s.Query(txid, q)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			if err = rows.Scan(&vID); err != nil {
				rows.Close()
				return nil, err
			}
			nodes[i].Vols = append(nodes[i].Vols, vID)
		}
		rows.Close()
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
			vl_max_encoding_group,
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

		if err = rows.Scan(&v.ID, &v.Stat, &v.Node, &v.Size, &v.MaxEG, &v.Speed); err != nil {
			return nil, err
		}

		vols = append(vols, v)
	}

	for i, v := range vols {
		q = fmt.Sprintf(
			`
			SELECT
				egv_encoding_group
			FROM
				encoding_group_volume
			WHERE
				egv_volume=%d
			`, v.ID,
		)

		var egID cmap.ID
		rows, err = s.Query(txid, q)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			if err = rows.Scan(&egID); err != nil {
				rows.Close()
				return nil, err
			}
			vols[i].EncGrps = append(vols[i].EncGrps, egID)
		}
		rows.Close()
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
		INSERT INTO cluster_job (clj_type, clj_state, clj_event_type, clj_event_time, clj_log)
		VALUES ('%d', '%d', '%d', '%s', '%s')
		`, job.Type, job.State, job.Event.Type, job.Event.TimeStamp, job.Log,
		)
	} else {
		q = fmt.Sprintf(
			`
		INSERT INTO cluster_job (clj_type, clj_state, clj_event_type, clj_event_affected, clj_event_time, clj_log)
		VALUES ('%d', '%d', '%d', '%s', '%s', '%s')
		`, job.Type, job.State, job.Event.Type, job.Event.AffectedEG, job.Event.TimeStamp, job.Log,
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
		affectedEG = "is NULL"
	} else {
		affectedEG = "=" + event.AffectedEG.String()
	}

	q := fmt.Sprintf(
		`
		UPDATE cluster_job
		SET clj_state=%d, clj_finished_at='%s'
		WHERE clj_event_type=%d AND clj_event_affected %s AND clj_state=%d
		`, cluster.Merged, cluster.TimeNow(), event.Type, affectedEG, cluster.Ready,
	)

	_, err := s.Execute(txid, q)
	return err
}

func (s *clusterStore) UpdateJob(txid repository.TxID, job *cluster.Job) error {
	q := fmt.Sprintf(
		`
		UPDATE cluster_job
		SET clj_state=%d, clj_scheduled_at='%s', clj_finished_at='%s', clj_log='%s'
		WHERE clj_id=%d
		`, job.State, job.ScheduledAt, job.FinishedAt, job.Log, job.ID,
	)

	_, err := s.Execute(txid, q)
	return err
}

func (s *clusterStore) ListJob() []string {
	l := make([]string, 0)

	q := fmt.Sprintf(
		`
		SELECT 
			clj_id,
			clj_type,
			clj_state,
			clj_event_type,
			ifnull (clj_event_affected, -1),
			clj_event_time,
			ifnull (clj_scheduled_at, ''),
			ifnull (clj_finished_at, ''),
			ifnull (clj_log, '')
		FROM cluster_job
		`,
	)

	rows, err := s.Query(repository.NotTx, q)
	if err != nil {
		return l
	}
	defer rows.Close()

	for rows.Next() {
		var j cluster.Job
		if err = rows.Scan(
			&j.ID, &j.Type, &j.State, &j.Event.Type, &j.Event.AffectedEG,
			&j.Event.TimeStamp, &j.ScheduledAt, &j.FinishedAt, &j.Log,
		); err != nil {
			return l
		}

		js := fmt.Sprintf(
			"[ID: %d] [Type: %s] [State: %s] [EventType: %s] [EventAffectedEG: %s]\n"+
				"[EventTime: %s]\n[ScheduledAt: %s]\n[FinishedAt: %s]\n[Log: %s]\n",
			j.ID.Int64(), j.Type.String(), j.State.String(), j.Event.Type.String(), j.Event.AffectedEG.String(), j.Event.TimeStamp.String(), j.ScheduledAt.String(), j.FinishedAt.String(), j.Log.String(),
		)

		l = append(l, js)
	}

	return l
}

func (s *clusterStore) RegisterVolume(txid repository.TxID, v *cmap.Volume) error {
	q := fmt.Sprintf(
		`
		INSERT INTO volume (vl_node, vl_status, vl_size, vl_encoding_group, vl_max_encoding_group, vl_speed)
		VALUES(%d, '%s', '%d', '%d', '%d', '%s')
		`, v.Node, cmap.VolPrepared, v.Size, 0, v.MaxEG, v.Speed,
	)

	r, err := s.Execute(txid, q)
	if err != nil {
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return err
	}
	v.ID = cmap.ID(id)

	return nil
}

func (s *clusterStore) FetchJob(txid repository.TxID) (*cluster.Job, error) {
	q := fmt.Sprintf(
		`
		SELECT clj_id, clj_event_type, clj_event_affected, clj_event_time
		FROM cluster_job
		WHERE clj_state=%d AND clj_type=%d
		ORDER BY clj_id LIMIT 1
		`, cluster.Ready, cluster.Batch,
	)

	j := &cluster.Job{
		Type:  cluster.Batch,
		State: cluster.Run,
	}
	var nullableEventAffected sql.NullInt64
	if err := s.QueryRow(txid, q).Scan(
		&j.ID, &j.Event.Type, &nullableEventAffected, &j.Event.TimeStamp,
	); err != nil {
		return nil, err
	}

	if nullableEventAffected.Valid {
		j.Event.AffectedEG = cmap.ID(nullableEventAffected.Int64)
	}

	q = fmt.Sprintf(
		`
		UPDATE cluster_job
		SET clj_state=%d
		WHERE clj_id=%d AND clj_state=%d
		`, cluster.Run, j.ID, cluster.Ready,
	)
	_, err := s.Execute(txid, q)

	return j, err
}

func (s *clusterStore) MakeNewEncodingGroup(txid repository.TxID, encGrp *cmap.EncodingGroup) error {
	// Make new encoding group.
	q := fmt.Sprintf(
		`
		INSERT INTO encoding_group (eg_status)
		VALUES ('%s')
		`, encGrp.Stat.String(),
	)
	r, err := s.Store.Execute(txid, q)
	if err != nil {
		return errors.Wrap(err, "failed to create encoding group")
	}
	egID, err := r.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to create encoding group")
	}

	// Register each volumes.
	for role, v := range encGrp.Vols {
		q = fmt.Sprintf(
			`
			INSERT INTO encoding_group_volume (egv_encoding_group, egv_volume, egv_role)
			VALUES ('%d', '%d', '%d')
			`, egID, v.Int64(), role,
		)
		_, err = s.Store.Execute(txid, q)
		if err != nil {
			return errors.Wrap(err, "failed to create volume in encoding group table")
		}

		q = fmt.Sprintf(
			`
			UPDATE volume
			SET vl_encoding_group=vl_encoding_group+1
			WHERE vl_id in ('%d')
			`, v.Int64(),
		)
		_, err = s.Store.Execute(txid, q)
		if err != nil {
			return errors.Wrap(err, "failed to increase encoding group counting in volume")
		}
	}

	return nil
}

func (s *clusterStore) UpdateChangedCMap(txid repository.TxID, m *cmap.CMap) error {
	for _, v := range m.Vols {
		q := fmt.Sprintf(
			`
		UPDATE volume
		SET vl_status='%s', vl_size=%d, vl_max_encoding_group=%d, vl_speed='%s'
		WHERE vl_id=%d
		`, v.Stat, v.Size, v.MaxEG, v.Speed, v.ID,
		)

		_, err := s.Execute(txid, q)
		if err != nil {
			return errors.Wrap(err, "failed to update volume")
		}
	}

	for _, n := range m.Nodes {
		// Update node status.
		q := fmt.Sprintf(
			`
			UPDATE node
			SET node_status='%s'
			WHERE node_id=%d
			`, n.Stat, n.ID,
		)
		_, err := s.Execute(txid, q)
		if err != nil {
			return errors.Wrap(err, "failed to update node")
		}

		if n.Stat == cmap.NodeAlive {
			continue
		}

		// If node is suspect, then update related volumes.
		if n.Stat == cmap.NodeSuspect {
			q := fmt.Sprintf(
				`
				UPDATE volume
				SET vl_status='%s'
				WHERE vl_node=%d AND (vl_status='%s' OR vl_status='%s')
				`, cmap.VolSuspect, n.ID, cmap.VolActive, cmap.VolPrepared,
			)

			_, err := s.Execute(txid, q)
			if err != nil {
				return errors.Wrap(err, "failed to update encoding group")
			}
			continue
		}

		// If node is faulty, then update related volumes.
		if n.Stat == cmap.NodeFaulty {
			q := fmt.Sprintf(
				`
				UPDATE volume
				SET vl_status='%s'
				WHERE vl_node=%d AND vl_status!='%s'
				`, cmap.VolFailed, n.ID, cmap.VolActive,
			)

			_, err := s.Execute(txid, q)
			if err != nil {
				return errors.Wrap(err, "failed to update encoding group")
			}
			continue
		}
	}

	q := fmt.Sprintf(
		`
		UPDATE encoding_group
		SET eg_status='%s'
		WHERE eg_status='%s' AND eg_id IN (
			SELECT egv_encoding_group
			FROM encoding_group_volume
			WHERE egv_volume IN (
				SELECT vl_id
				FROM volume
				WHERE vl_status='%s' OR vl_status='%s'
			)
		)
		`, cmap.EGRdonly, cmap.EGAlive, cmap.VolSuspect, cmap.VolFailed,
	)
	_, err := s.Execute(txid, q)
	if err != nil {
		return errors.Wrap(err, "failed to update encoding group")
	}

	return nil
}
