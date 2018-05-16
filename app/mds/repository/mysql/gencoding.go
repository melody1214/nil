package mysql

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/gencoding"
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
		INSERT INTO global_encoding_group (geg_region_frst, geg_region_secd, geg_region_thrd, geg_region_four, geg_state)
		SELECT * FROM (SELECT '%d', '%d', '%d', '%d', '%d') AS tmp
		WHERE NOT EXISTS (
			SELECT geg_id
			FROM global_encoding_group
			WHERE geg_region_frst = '%d' AND geg_region_secd = '%d' AND geg_region_thrd = '%d' AND geg_region_four = '%d'
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

func (s *gencodingStore) Make() error {
	q := fmt.Sprintf(
		`
		SELECT
			ger_region, ger_encoding_group_chunk
		FROM
			global_encoding_request
		ORDER BY ger_encoding_group_chunk DESC limit 4
		`,
	)

	rows, err := s.Query(repository.NotTx, q)
	if err != nil {
		return err
	}
	defer rows.Close()

	type ger struct {
		region int
		chunk  int
	}

	rs := make([]ger, 0)
	for rows.Next() {
		var r ger

		if err = rows.Scan(&r); err != nil {
			return err
		}

		rs = append(rs, r)
	}

	if len(rs) != 4 {
		return fmt.Errorf("not enough request information")
	}

	for _, r := range rs {
		if r.chunk == 0 {
			return fmt.Errorf("not enough chunk to encode")
		}
	}

	var gegID = 0
	for i := 0; i < 20; i++ {
		randIdx := s.random.Perm(len(rs))

		q = fmt.Sprintf(
			`
		SELECT
			geg_id
		FROM
			global_encoding_group
		WHERE
			geg_region_frst = %d AND geg_region_secd = %d AND geg_region_thrd = %d AND geg_region_four = %d
		`, rs[randIdx[0]].region, rs[randIdx[1]].region, rs[randIdx[2]].region, rs[randIdx[3]].region,
		)

		if err := s.QueryRow(repository.NotTx, q).Scan(&gegID); err != nil {
			continue
		}
	}

	if gegID == 0 {
		return fmt.Errorf("failed to find the global encoding group with the selected regions")
	}

	q = fmt.Sprintf(
		`
		INSERT INTO global_encoding_table (get_global_encoding_group, get_status)
		VALUES ('%d', '%d')
		`, gegID, 0,
	)

	_, err = s.PublishCommand("execute", q)
	return err
}
