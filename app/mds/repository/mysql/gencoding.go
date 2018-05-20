package mysql

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/gencoding"
	"github.com/chanyoung/nil/app/mds/usecase/gencoding/token"
	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
)

type gencodingStore struct {
	*Store
	random *rand.Rand
}

// NewGencodingRepository returns a new instance of a gencoding repository.
func NewGencodingRepository(s *Store) gencoding.Repository {
	return &gencodingStore{
		Store:  s,
		random: rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// GenerateGencodingGroup generates a global encoding group with the given regions list.
func (s *gencodingStore) GenerateGencodingGroup(regions []string) error {
	if len(regions) != 4 {
		return fmt.Errorf("region number is must four")
	}

	regionIDs := make([]int, len(regions))
	for i := 0; i < len(regions); i++ {
		q := fmt.Sprintf(
			`
			SELECT rg_id FROM region WHERE rg_name='%s'
			`, regions[i],
		)

		if err := s.QueryRow(repository.NotTx, q).Scan(&regionIDs[i]); err != nil {
			return errors.Wrapf(err, "failed to find region id with the given name: %s", regions[i])
		}
	}

	q := fmt.Sprintf(
		`
		INSERT INTO global_encoding_group (geg_region_first, geg_region_second, geg_region_third, geg_region_parity, geg_state)
		SELECT * FROM (SELECT '%d', '%d', '%d', '%d', '%d') AS tmp
		WHERE NOT EXISTS (
			SELECT geg_id
			FROM global_encoding_group
			WHERE geg_region_first = '%d' AND geg_region_second = '%d' AND geg_region_third = '%d' AND geg_region_parity = '%d'
		) LIMIT 1
		`, regionIDs[0], regionIDs[1], regionIDs[2], regionIDs[3], 0, regionIDs[0], regionIDs[1], regionIDs[2], regionIDs[3],
	)

	_, err := s.PublishCommand("execute", q)
	if err == nil {
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
		return fmt.Errorf("duplicated entry")
	default:
		return err
	}
}

func (s *gencodingStore) AmILeader() bool {
	return s.raft.State() == raft.Leader
}

// MakeGlobalEncodingJob creates the global encoding job by the given token information.
func (s *gencodingStore) MakeGlobalEncodingJob(t *token.Token, p *token.Unencoded) error {
	// Make global encoding job.
	q := fmt.Sprintf(
		`
		INSERT INTO global_encoding_job (gej_status)
		VALUES (%d)
		`, gencoding.Ready,
	)
	r, err := s.PublishCommand("execute", q)
	if err != nil {
		return err
	}
	gblChunkID, err := r.LastInsertId()
	if err != nil {
		return err
	}

	rollback := false
	// Register unencoded chunks.
	unencodedChunks := [3]token.Unencoded{t.First, t.Second, t.Third}
	for i, c := range unencodedChunks {
		q = fmt.Sprintf(
			`
			INSERT INTO global_encoding_chunk (guc_job, guc_role, guc_region, guc_node, guc_volume, guc_encgrp, guc_chunk)
			VALUES ('%d', '%d', '%d', '%d', '%d', '%d', '%s')
			`, gblChunkID, i, c.Region.RegionID, c.Node, c.Volume, c.EncGrp, c.ChunkID,
		)
		_, err = s.PublishCommand("execute", q)
		if err != nil {
			rollback = true
			break
		}
	}

	// Add primary information.
	q = fmt.Sprintf(
		`
		INSERT INTO global_encoding_chunk (guc_job, guc_role, guc_region, guc_node, guc_volume, guc_encgrp, guc_chunk)
		VALUES ('%d', '%d', '%d', '%d', '%d', '%d', '%d')
		`, gblChunkID, 3, p.Region.RegionID, p.Node, p.Volume, p.EncGrp, gblChunkID,
	)
	_, err = s.PublishCommand("execute", q)
	if err != nil {
		rollback = true
	}

	// Check if errors are occured.
	if rollback {
		q = fmt.Sprintf(
			`
			DELETE FROM global_encoding_chunk
			WHERE guc_job=%d
			`, gblChunkID,
		)
		s.PublishCommand("execute", q)
		q = fmt.Sprintf(
			`
			DELETE FROM global_encoding_job
			WHERE gej_id=%d
			`, gblChunkID,
		)
		s.PublishCommand("execute", q)
	}

	return nil
}

func (s *gencodingStore) LeaderEndpoint() (endpoint string) {
	if s.raft == nil {
		return
	}

	future := s.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return
	}

	servers := future.Configuration().Servers
	// Not joined yet.
	if len(servers) == 1 {
		return
	}

	var leader *raft.Server
	leaderAddress := s.raft.Leader()
	for _, s := range servers {
		if s.Address == leaderAddress {
			leader = &s
			break
		}
	}

	if leader == nil {
		return
	}

	q := fmt.Sprintf(
		`
		SELECT rg_end_point
		FROM region
		WHERE rg_name='%s'
		`, string(leader.ID),
	)

	s.QueryRow(repository.NotTx, q).Scan(&endpoint)
	return
}

func (s *gencodingStore) GetJob(regionName string) (*token.Token, error) {
	q := fmt.Sprintf(
		`
		SELECT rg_id
		FROM region
		WHERE rg_name='%s'
		`, regionName,
	)

	var regionID int64
	if err := s.QueryRow(repository.NotTx, q).Scan(&regionID); err != nil {
		return nil, err
	}

	q = fmt.Sprintf(
		`
		SELECT guc_job
		FROM global_encoding_chunk
		WHERE guc_role=%d AND guc_region=%d
		`, 3, regionID,
	)

	rows, err := s.Query(repository.NotTx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobID int64
	for rows.Next() {
		if err := rows.Scan(&jobID); err != nil {
			return nil, err
		}

		q = fmt.Sprintf(
			`
			UPDATE global_encoding_job
			SET gej_status=IF(gej_status=%d,%d,gej_status)
			WHERE gej_id=%d
			`, gencoding.Ready, gencoding.Run, jobID,
		)

		r, err := s.PublishCommand("execute", q)
		if err != nil {
			return nil, err
		}
		if affected, err := r.RowsAffected(); err != nil {
			return nil, err
		} else if affected == 1 {
			break
		}
		jobID = 0
	}

	if jobID == 0 {
		return nil, fmt.Errorf("no jobs for you")
	}

	t := &token.Token{}

	q = fmt.Sprintf(
		`
		SELECT guc_role, guc_region, guc_node, guc_volume, guc_encgrp, guc_chunk
		FROM global_encoding_chunk
		WHERE guc_job=%d
		`, jobID,
	)
	rows, err = s.Query(repository.NotTx, q)
	if err != nil {
		q = fmt.Sprintf(
			`
			UPDATE global_encoding_job
			SET gej_status=%d
			WHERE gej_id=%d
			`, gencoding.Fail, jobID,
		)
		s.PublishCommand("execute", q)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		role := 0
		var c token.Unencoded
		if err := rows.Scan(&role, &c.Region.RegionID, &c.Node, &c.Volume, &c.EncGrp, &c.ChunkID); err != nil {
			q = fmt.Sprintf(
				`
				UPDATE global_encoding_job
				SET gej_status=%d
				WHERE gej_id=%d
				`, gencoding.Fail, jobID,
			)
			s.PublishCommand("execute", q)
			return nil, err
		}

		q = fmt.Sprintf(
			`
			SELECT rg_name, rg_end_point
			FROM region
			WHERE rg_id=%d
			`, c.Region.RegionID,
		)
		if err := s.QueryRow(repository.NotTx, q).Scan(&c.Region.RegionName, &c.Region.Endpoint); err != nil {
			return nil, err
		}

		if role == 0 {
			t.First = c
		} else if role == 1 {
			t.Second = c
		} else if role == 2 {
			t.Third = c
		} else if role == 3 {
			t.Primary = c
		} else {
			q = fmt.Sprintf(
				`
				UPDATE global_encoding_job
				SET gej_status=%d
				WHERE gej_id=%d
				`, gencoding.Fail, jobID,
			)
			s.PublishCommand("execute", q)
			return nil, fmt.Errorf("unknown encoding chunk role")
		}
	}

	return t, nil
}

func (s *gencodingStore) RegionEndpoint(regionID int) (endpoint string) {
	q := fmt.Sprintf(
		`
		SELECT rg_end_point
		FROM region
		WHERE rg_id='%d'
		`, regionID,
	)

	s.QueryRow(repository.NotTx, q).Scan(&endpoint)
	return
}

func (s *gencodingStore) GetRoutes(leaderEndpoint string) (*token.Leg, error) {
	q := fmt.Sprintf(
		`
		SELECT rg_id, rg_name, rg_end_point FROM region
		`,
	)

	rows, err := s.Query(repository.NotTx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	l := &token.Leg{
		CurrentIdx: 0,
		Stops:      make([]token.Stop, 0),
	}

	for rows.Next() {
		s := token.Stop{}

		if err := rows.Scan(&s.RegionID, &s.RegionName, &s.Endpoint); err != nil {
			return nil, err
		}

		l.Stops = append(l.Stops, s)
	}

	for i, s := range l.Stops {
		if string(s.Endpoint) != leaderEndpoint {
			continue
		}

		// Swap to make leader come to the first.
		l.Stops[i] = l.Stops[0]
		l.Stops[0] = s
	}

	return l, nil
}

func (s *gencodingStore) SetJobStatus(id int64, status gencoding.Status) error {
	q := fmt.Sprintf(
		`
		UPDATE global_encoding_job
		SET gej_status=%d
		WHERE gej_id=%d
		`, status, id,
	)

	_, err := s.PublishCommand("execute", q)
	return err
}

func (s *gencodingStore) JobFinished(t *token.Token) error {
	q := fmt.Sprintf(
		`
		SELECT geg_id
		FROM global_encoding_group
		WHERE geg_region_first=%d AND geg_region_second=%d AND geg_region_third=%d AND geg_region_parity=%d
		`, t.First.Region.RegionID, t.Second.Region.RegionID, t.Third.Region.RegionID, t.Primary.Region.RegionID,
	)
	var gblEncodingGroupID int64
	if err := s.QueryRow(repository.NotTx, q).Scan(&gblEncodingGroupID); err != nil {
		return errors.Wrap(err, "failed to find global encoding group id")
	}

	q = fmt.Sprintf(
		`
		INSERT INTO global_encoded_chunk (gec_chunk_id, gec_global_encoding_group, gec_local_encoding_group_first, gec_local_encoding_group_second, gec_local_encoding_group_third, gec_local_encoding_group_parity)
		VALUES (%s, %d, %d, %d, %d, %d)
		`, t.Primary.ChunkID, gblEncodingGroupID, t.First.EncGrp, t.Second.EncGrp, t.Third.EncGrp, t.Primary.EncGrp,
	)
	_, err := s.PublishCommand("execute", q)
	if err != nil {
		return errors.Wrap(err, "failed to insert global encoding chunk")
	}

	q = fmt.Sprintf(
		`
		UPDATE global_encoding_job
		SET gej_status=%d
		WHERE gej_id=%s
		`, gencoding.Done, t.Primary.ChunkID,
	)
	_, err = s.PublishCommand("execute", q)
	if err != nil {
		return errors.Wrap(err, "failed to update job status to done")
	}

	return nil
}

func (s *gencodingStore) RemoveFailedJobs() error {
	q := fmt.Sprintf(
		`
		SELECT gej_id
		FROM global_encoding_job
		WHERE gej_status=%d
		`, gencoding.Fail,
	)
	rows, err := s.Query(repository.NotTx, q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var jobID int64
	for rows.Next() {
		if err := rows.Scan(&jobID); err != nil {
			return err
		}

		q = fmt.Sprintf(
			`
			DELETE FROM global_encoding_chunk
			WHERE guc_job=%d
			`, jobID,
		)
		_, err = s.PublishCommand("execute", q)
		if err != nil {
			return err
		}

		q = fmt.Sprintf(
			`
			DELETE FROM global_encoding_job
			WHERE gej_id=%d
			`, jobID,
		)
		_, err = s.PublishCommand("execute", q)
		if err != nil {
			return err
		}
	}
	return nil
}
