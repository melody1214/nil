package membership

import "github.com/pkg/errors"
import "github.com/sirupsen/logrus"

var logger *logrus.Entry

// Service is the root manager of membership package.
// The service consists of four parts described below.
//
// 1. Server
// The server is a membership management server based on the swim membership
// protocol. It sends new updates to randomly selected nodes and updates its
// membership information. This server is designed as a finite state machine
// and does not process jobs with threaded way, because I thought if the
// membership server failed to handle requests in next ping period, it implies
// there is too much burden for membership server.
//
// 2. Cluster map
// Cluster map contains the information of each node, volume, encoding group
// and etc. It is versioned for every significant changes are occurred.
//
// 3. Slave cluster map api
// Slave cluster map api provides functions to search various elements of
// cluster map with the given conditions. It also provides the functions
// to update volume status or capacity information.
//
// 4. Master cluster map api
// Master cluster map api is the superset of slave api. Additionally provides add,
// remove, update node functions and all of this kind of changes will increment
// the version number of the cluster map.
type Service struct {
	cfg         Config
	cMapManager *cMapManager
}

// NewService returns new membership service.
func NewService(cfg Config, log *logrus.Entry) (*Service, error) {
	logger = log

	s := &Service{
		cfg: cfg,
	}

	cm, err := newCMapManager(cfg.Coordinator)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cmap manager")
	}

	s.cMapManager = cm
	return s, nil
}

// SlaveAPI is the interface for access the membership service with slave mode.
type SlaveAPI interface {
}

// SlaveAPI returns a set of APIs that can be used by nodes in slave mode.
func (s *Service) SlaveAPI() SlaveAPI {
	return s
}

// MasterAPI is the interface for access the membership service with master mode.
type MasterAPI interface {
}

// MasterAPI returns a set of APIs that can be used by nodes in master mode.
func (s *Service) MasterAPI() MasterAPI {
	return s
}
