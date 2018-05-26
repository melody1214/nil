package cmap

import (
	"fmt"

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

	// Managing cmap with membership server and client APIs.
	manager *manager

	// Membership protocol server.
	server *server
}

// NewService returns new membership service.
func NewService(coordinator NodeAddress, log *logrus.Entry) (*Service, error) {
	logger = log

	cm, err := newManager(coordinator)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cmap manager")
	}

	return &Service{
		manager: cm,
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
	UpdateVolume(volume Volume) error
	UpdateUnencoded(egID ID, unencoded int) error
}

// MasterAPI returns a set of APIs that can be used by nodes in master mode.
func (s *Service) MasterAPI() MasterAPI {
	return s
}

// SlaveAPI returns a set of APIs that can be used by nodes in slave mode.
func (s *Service) SlaveAPI() SlaveAPI {
	return s
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

// UpdateVolume updates the volume status of the given volume ID.
func (s *Service) UpdateVolume(volume Volume) error {
	node, err := s.SearchCall().Node().ID(volume.Node).Do()
	if err != nil {
		return fmt.Errorf("no such node: %v", err)
	}
	if node.Name != s.cfg.Name {
		return fmt.Errorf("only can update volumes which this node has")
	}

	s.manager.UpdateVolume(volume)
	return nil
}

// UpdateUnencoded updates the unencoded field of encoding group.
func (s *Service) UpdateUnencoded(egID ID, unencoded int) error {
	c := s.SearchCall()
	eg, err := c.EncGrp().ID(egID).Do()
	if err != nil {
		return errors.Wrap(err, "failed to find encoding group with the given id")
	}
	node, err := c.Node().ID(eg.LeaderVol()).Do()
	if err != nil {
		return errors.Wrap(err, "failed to find leader node with the given encoding group")
	}
	if node.Name != s.cfg.Name {
		return fmt.Errorf("only can update eg which this the leader volume")
	}

	s.manager.UpdateUnencoded(egID, unencoded)
	return nil
}
