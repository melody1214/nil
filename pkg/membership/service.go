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
	SearchCallNode() *SearchCallNode
	SearchCallVolume() *SearchCallVolume
	SearchCallEncGrp() *SearchCallEncGrp
	UpdateNodeStatus(nID ID, stat NodeStatus) error
	UpdateVolumeStatus(vID ID, stat VolumeStatus) error
	UpdateEncodingGroupStatus(egID ID, stat EncodingGroupStatus) error
	UpdateEncodingGroupUsed(egID ID, used uint64) error
	GetUpdatedNoti(ver CMapVersion) <-chan interface{}
}

// SlaveAPI returns a set of APIs that can be used by nodes in slave mode.
func (s *Service) SlaveAPI() SlaveAPI {
	return s
}

// MasterAPI is the interface for access the membership service with master mode.
type MasterAPI interface {
	SearchCallNode() *SearchCallNode
	SearchCallVolume() *SearchCallVolume
	SearchCallEncGrp() *SearchCallEncGrp
	UpdateCMap(cmap CMap) error
	GetOutdatedNoti() <-chan interface{}
	GetUpdatedNoti(ver CMapVersion) <-chan interface{}
}

// MasterAPI returns a set of APIs that can be used by nodes in master mode.
func (s *Service) MasterAPI() MasterAPI {
	return s
}

// UpdateNodeStatus updates the node status of the given node ID.
func (s *Service) UpdateNodeStatus(nID ID, stat NodeStatus) error {
	return nil
}

// UpdateVolumeStatus updates the volume status of the given volume ID.
func (s *Service) UpdateVolumeStatus(vID ID, stat VolumeStatus) error {
	return nil
}

// UpdateEncodingGroupStatus updates the status of encoding group.
func (s *Service) UpdateEncodingGroupStatus(egID ID, stat EncodingGroupStatus) error {
	return nil
}

// UpdateEncodingGroupUsed updates the used size of encoding group.
func (s *Service) UpdateEncodingGroupUsed(egID ID, used uint64) error {
	return nil
}

// UpdateCMap updates the new cmap manager with the given cmap.
func (s *Service) UpdateCMap(cmap CMap) error {
	return nil
}

// SearchCallNode returns a new search call for finding node.
func (s *Service) SearchCallNode() *SearchCallNode {
	return s.cMapManager.SearchCallNode()
}

// SearchCallVolume returns a new search call for finding volume.
func (s *Service) SearchCallVolume() *SearchCallVolume {
	return s.cMapManager.SearchCallVolume()
}

// SearchCallEncGrp returns a new search call for finding encoding group.
func (s *Service) SearchCallEncGrp() *SearchCallEncGrp {
	return s.cMapManager.SearchCallEncGrp()
}

// GetOutdatedNoti returns a channel which will send notification when
// the cluster map is outdated.
func (s *Service) GetOutdatedNoti() <-chan interface{} {
	return s.cMapManager.GetOutdatedNoti()
}

// GetUpdatedNoti returns a channel which will send notification when
// the higher version of cluster map is created.
func (s *Service) GetUpdatedNoti(ver CMapVersion) <-chan interface{} {
	return s.cMapManager.GetUpdatedNoti(ver)
}
