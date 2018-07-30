package mysql

import (
	"fmt"

	raftdomain "github.com/chanyoung/nil/app/mds/domain/service/raft"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/hashicorp/raft"
)

type raftSimpleService struct {
	rs *raftService
}

func (s *raftService) NewRaftSimpleService() raftdomain.SimpleService {
	return &raftSimpleService{
		rs: s,
	}
}

func (ss *raftSimpleService) Leader() (bool, error) {
	if !ss.rs.opened {
		return false, ErrRaftNotOpened
	}

	return ss.rs.raft.State() == raft.Leader, nil
}

func (ss *raftSimpleService) LeaderEndPoint() (string, error) {
	ctxLogger := mlog.GetMethodLogger(logger, "raftSimpleService.LeaderEndPoint")

	if !ss.rs.opened {
		return "", ErrRaftNotOpened
	}

	future := ss.rs.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		ctxLogger.Error(err)
		return "", ErrRaftInternal
	}

	servers := future.Configuration().Servers
	if len(servers) == 1 {
		ctxLogger.Error("number of raft joined server is 1")
		return "", ErrRaftInternal
	}

	var leader *raft.Server
	leaderAddr := ss.rs.raft.Leader()
	for _, s := range servers {
		if s.Address == leaderAddr {
			leader = &s
			break
		}
	}

	if leader == nil {
		ctxLogger.Errorf("no leader node in the server list: %v", servers)
		return "", ErrRaftInternal
	}

	q := fmt.Sprintf(
		`
        SELECT rg_end_point
        FROM region
        WHERE rg_name='%s'
        `, string(leader.ID),
	)

	var endPoint string
	ss.rs.store.QueryRow(repository.NotTx, q).Scan(&endPoint)

	return endPoint, nil
}
