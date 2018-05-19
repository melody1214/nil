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
			fmt.Printf("%+v\n\n", err)
			break
		}
	}

	// Add primary information.
	q = fmt.Sprintf(
		`
		INSERT INTO global_encoding_chunk (guc_job, guc_role, guc_region, guc_node, guc_volume, guc_encgrp, guc_chunk)
		VALUES ('%d', '%d', '%d', '%d', '%d', '%d', '%s')
		`, gblChunkID, 3, p.Region.RegionID, p.Node, p.Volume, p.EncGrp, "",
	)
	_, err = s.PublishCommand("execute", q)
	if err != nil {
		fmt.Printf("%+v\n\n", err)
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

// func (s *gencodingStore) Make() (*gencoding.Table, error) {
// 	if s.raft == nil {
// 		return nil, fmt.Errorf("raft is nil")
// 	}

// 	q := fmt.Sprintf(
// 		`
// 		SELECT
// 			ger_region, ger_encoding_group_chunk
// 		FROM
// 			global_encoding_request
// 		ORDER BY ger_encoding_group_chunk DESC
// 		`,
// 	)

// 	rows, err := s.Query(repository.NotTx, q)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	type ger struct {
// 		region int
// 		chunk  int
// 	}

// 	rs := make([]ger, 0)
// 	for rows.Next() {
// 		var r ger

// 		if err = rows.Scan(&r.region, &r.chunk); err != nil {
// 			return nil, err
// 		}

// 		rs = append(rs, r)
// 	}

// 	if len(rs) < 4 {
// 		return nil, fmt.Errorf("not enough request information")
// 	}

// 	for i, r := range rs {
// 		if i < 3 && r.chunk == 0 {
// 			return nil, fmt.Errorf("not enough chunk to encode")
// 		}
// 	}

// 	tbl := &gencoding.Table{
// 		RegionIDs: make([]int, 4),
// 		Status:    gencoding.Ready,
// 	}

// 	var gegID = 0
// 	for i := 0; i < 20; i++ {
// 		randIdx := s.random.Perm(len(rs) - 3)

// 		q = fmt.Sprintf(
// 			`
// 		SELECT
// 			geg_id
// 		FROM
// 			global_encoding_group
// 		WHERE
// 			geg_region_first = %d AND geg_region_second = %d AND geg_region_third = %d AND geg_region_parity = %d
// 		`, rs[0].region, rs[1].region, rs[2].region, rs[3+randIdx[0]].region,
// 		)

// 		if err := s.QueryRow(repository.NotTx, q).Scan(&gegID); err != nil {
// 			continue
// 		}

// 		tbl.RegionIDs[0] = rs[0].region
// 		tbl.RegionIDs[1] = rs[1].region
// 		tbl.RegionIDs[2] = rs[2].region
// 		tbl.RegionIDs[3] = rs[3+randIdx[0]].region
// 		break
// 	}

// 	if gegID == 0 {
// 		return nil, fmt.Errorf("failed to find the global encoding group with the selected regions")
// 	}

// 	q = fmt.Sprintf(
// 		`
// 		INSERT INTO global_encoding_table (get_global_encoding_group, get_status)
// 		VALUES ('%d', '%d')
// 		`, gegID, gencoding.Ready,
// 	)
// 	r, err := s.PublishCommand("execute", q)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tbl.ID, err = r.LastInsertId()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return tbl, err
// }

// func (s *gencodingStore) UpdateUnencodedChunks(regionName string, unencoded int) error {
// 	q := fmt.Sprintf(
// 		`
// 		SELECT rg_id
// 		FROM region
// 		WHERE rg_name='%s'
// 		`, regionName,
// 	)
// 	var regionID int
// 	if err := s.QueryRow(repository.NotTx, q).Scan(&regionID); err != nil {
// 		return err
// 	}

// 	q = fmt.Sprintf(
// 		`
// 		INSERT INTO global_encoding_request (ger_region, ger_encoding_group_chunk)
// 		VALUES (%d, %d)
// 		ON DUPLICATE KEY UPDATE ger_region=%d, ger_encoding_group_chunk=%d
// 		`, regionID, unencoded, regionID, unencoded,
// 	)

// 	_, err := s.PublishCommand("execute", q)
// 	return err
// }

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
