package cmap

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Service is the root manager of membership package.
// The service consists of four parts described below.
//
// 1. Server
// The server is a membership management server based on the swim membership
// protocol. It sends new updates to randomly selected nodes and updates its
// membership information.
//
// 2. Cluster map
// Cluster map contains the information of each node, volume, encoding group
// and etc. It is versioned for every significant changes are occurred.
//
// 3. Slave cluster map api
// Slave cluster map api provides functions to search various elements of
// cluster map with the given conditions. It also provides the functions
// to update volume status or capacity information. (for DS functions)
//
// 4. Master cluster map api
// Master cluster map api is the superset of slave api. Additionally provides add,
// remove, update node functions and all of this kind of changes will increment
// the version number of the cluster map. (for MDS functions)
type Service struct {
	// Configuration provided at service initialization.
	cfg Config

	// id represents this node's id.
	id ID

	// Managing cmap with membership server and client APIs.
	manager *manager

	// Membership protocol server.
	server *server
}

// NewService returns new membership service.
func NewService(log *logrus.Entry) (*Service, error) {
	logger = log

	return &Service{
		manager: newManager(),
		id:      ID(-1),
	}, nil
}

// StartMembershipServer starts membership server to gossip.
func (s *Service) StartMembershipServer(cfg Config, trans Transport) error {
	s.cfg = cfg
	swimSrv, err := newServer(cfg, s.manager, trans)
	if err != nil {
		return errors.Wrap(err, "failed to make new swim server")
	}
	s.server = swimSrv

	go s.server.run()
	return nil
}

// CommonAPI is the interface for access the membership service with any mode.
type CommonAPI interface {
	SearchCall() *SearchCall
	GetStateChangedNoti() <-chan interface{}
	GetUpdatedNoti(ver Version) <-chan interface{}
	ID() (ID, error)
}

// MasterAPI is the interface for access the membership service with master mode.
type MasterAPI interface {
	CommonAPI
	GetLatestCMap() CMap
	UpdateCMap(cmap *CMap) error
}

// SlaveAPI is the interface for access the membership service with slave mode.
type SlaveAPI interface {
	CommonAPI
}

// MasterAPI returns a set of APIs that can be used by nodes in master mode.
func (s *Service) MasterAPI() MasterAPI {
	return s
}

// SlaveAPI returns a set of APIs that can be used by nodes in slave mode.
func (s *Service) SlaveAPI() SlaveAPI {
	return s
}

// ErrNotInitialized is used when the cmap API is called before the service is initialized.
var ErrNotInitialized = errors.New("cmap service is not initialized yet")

// ID returns this node ID.
func (s *Service) ID() (ID, error) {
	if s.id.Int64() > 0 {
		return s.id, nil
	}

	// ID is not initialized.
	n, err := s.SearchCall().Node().Name(s.cfg.Name).Do()
	if err != nil {
		return s.id, ErrNotInitialized
	}

	// Save the id.
	s.id = n.ID

	return s.id, nil
}

// SearchCall returns a SearchCall object which can support convenient
// searching some members in the cluster.
func (s *Service) SearchCall() *SearchCall {
	return s.manager.SearchCall()
}

// GetStateChangedNoti returns a channel which will send notification when
// the cluster map is outdated.
func (s *Service) GetStateChangedNoti() <-chan interface{} {
	return s.manager.GetStateChangedNoti()
}

// GetUpdatedNoti returns a channel which will send notification when
// the higher version of cluster map is created.
func (s *Service) GetUpdatedNoti(ver Version) <-chan interface{} {
	return s.manager.GetUpdatedNoti(ver)
}

// GetLatestCMap returns the latest cluster map.
func (s *Service) GetLatestCMap() CMap {
	return *s.manager.LatestCMap()
}

// UpdateCMap updates the new cmap manager with the given cmap.
func (s *Service) UpdateCMap(cmap *CMap) error {
	s.manager.mergeCMap(cmap)
	return nil
}
