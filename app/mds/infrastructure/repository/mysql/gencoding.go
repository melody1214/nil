package mysql

import (
	"math/rand"
	"time"

	"github.com/chanyoung/nil/app/mds/application/gencoding"
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

// func (s *gencodingStore) AmILeader() (bool, error) {
// 	if s.raft == nil {
// 		return false, fmt.Errorf("raft is not initialized yet")
// 	}
// 	return s.raft.State() == raft.Leader, nil
// }

// func (s *gencodingStore) LeaderEndpoint() (endpoint string) {
// 	return s.leaderEndPoint()
// }

// func (s *gencodingStore) RegionEndpoint(regionID int) (endpoint string) {
// 	q := fmt.Sprintf(
// 		`
// 		SELECT rg_end_point
// 		FROM region
// 		WHERE rg_id='%d'
// 		`, regionID,
// 	)

// 	s.QueryRow(repository.NotTx, q).Scan(&endpoint)
// 	return
// }
