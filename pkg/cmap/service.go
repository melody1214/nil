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
	cfg         Config
	cMapManager *cMapManager
	server      *server
}

// NewService returns new membership service.
func NewService(coordinator NodeAddress, log *logrus.Entry) (*Service, error) {
	logger = log

	cm, err := newCMapManager(coordinator)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cmap manager")
	}

	return &Service{
		cMapManager: cm,
	}, nil
}

// StartMembershipServer starts membership server to gossip.
func (s *Service) StartMembershipServer(cfg Config, trans Transport) error {
	s.cfg = cfg
	swimSrv, err := newServer(cfg, s.cMapManager, trans)
	if err != nil {
		return errors.Wrap(err, "failed to make new swim server")
	}
	s.server = swimSrv

	go s.server.run()
	return nil
}

// SlaveAPI is the interface for access the membership service with slave mode.
type SlaveAPI interface {
	SearchCallNode() *SearchCallNode
	SearchCallVolume() *SearchCallVolume
	SearchCallEncGrp() *SearchCallEncGrp
	UpdateNodeStatus(nID ID, stat NodeStatus) error
	UpdateVolume(volume Volume) error
	UpdateEncodingGroupStatus(egID ID, stat EncodingGroupStatus) error
	UpdateEncodingGroupUsed(egID ID, used uint64) error
	GetLatestCMapVersion() Version
	GetLatestCMap() CMap
	GetUpdatedNoti(ver Version) <-chan interface{}
	FindEncodingGroupByLeader(leaderNode ID) []EncodingGroup
	UpdateEncodingGroupUnencoded(eg EncodingGroup) error
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
	GetLatestCMap() CMap
	UpdateCMap(cmap *CMap) error
	GetStateChangedNoti() <-chan interface{}
	GetLatestCMapVersion() Version
	GetUpdatedNoti(ver Version) <-chan interface{}
}

// MasterAPI returns a set of APIs that can be used by nodes in master mode.
func (s *Service) MasterAPI() MasterAPI {
	return s
}

// UpdateNodeStatus updates the node status of the given node ID.
func (s *Service) UpdateNodeStatus(nID ID, stat NodeStatus) error {
	return nil
}

// UpdateVolume updates the volume status of the given volume ID.
func (s *Service) UpdateVolume(volume Volume) error {
	node, err := s.cMapManager.SearchCallNode().ID(volume.Node).Do()
	if err != nil {
		return fmt.Errorf("no such node: %v", err)
	}
	if node.Name != s.cfg.Name {
		return fmt.Errorf("only can update volumes which this node has")
	}

	s.cMapManager.mu.Lock()
	defer s.cMapManager.mu.Unlock()

	cm := s.cMapManager.latestCMap()
	for i, v := range cm.Vols {
		if v.ID != volume.ID {
			continue
		}

		cm.Vols[i].Stat = volume.Stat
		cm.Vols[i].Size = volume.Size
		cm.Vols[i].Speed = volume.Speed
		cm.Vols[i].Incr = cm.Vols[i].Incr + 1
	}

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

// UpdateEncodingGroupUnencoded updates the unencoded field of encoding group.
func (s *Service) UpdateEncodingGroupUnencoded(eg EncodingGroup) error {
	vol, err := s.cMapManager.SearchCallVolume().ID(eg.Vols[len(eg.Vols)-1]).Do()
	if err != nil {
		return fmt.Errorf("no such volume: %v", err)
	}
	node, err := s.cMapManager.SearchCallNode().ID(vol.Node).Do()
	if node.Name != s.cfg.Name {
		return fmt.Errorf("only can update eg which this the leader volume")
	}

	s.cMapManager.mu.Lock()
	defer s.cMapManager.mu.Unlock()

	cm := s.cMapManager.latestCMap()
	for i, found := range cm.EncGrps {
		if found.ID != eg.ID {
			continue
		}

		cm.EncGrps[i].Uenc = eg.Uenc
		cm.EncGrps[i].Incr = cm.EncGrps[i].Incr + 1
	}

	return nil
}

// GetLatestCMap returns the latest cluster map.
func (s *Service) GetLatestCMap() CMap {
	return s.cMapManager.LatestCMap()
}

// GetLatestCMapVersion returns the latest version of cluster map.
func (s *Service) GetLatestCMapVersion() Version {
	return s.cMapManager.latest
}

// UpdateCMap updates the new cmap manager with the given cmap.
func (s *Service) UpdateCMap(cmap *CMap) error {
	s.cMapManager.mergeCMap(cmap)
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

// GetStateChangedNoti returns a channel which will send notification when
// the cluster map is outdated.
func (s *Service) GetStateChangedNoti() <-chan interface{} {
	return s.cMapManager.GetStateChangedNoti()
}

// GetUpdatedNoti returns a channel which will send notification when
// the higher version of cluster map is created.
func (s *Service) GetUpdatedNoti(ver Version) <-chan interface{} {
	return s.cMapManager.GetUpdatedNoti(ver)
}

// FindEncodingGroupByLeader finds encoding groups owned by given leader node.
func (s *Service) FindEncodingGroupByLeader(leaderNode ID) []EncodingGroup {
	m := s.cMapManager.LatestCMap()

	vmap := make(map[ID]Volume, 0)
	for _, v := range m.Vols {
		if v.Node == leaderNode {
			vmap[v.ID] = v
		}
	}

	egs := make([]EncodingGroup, 0)
	for _, eg := range m.EncGrps {
		if _, ok := vmap[eg.Vols[len(eg.Vols)-1]]; ok {
			egs = append(egs, eg)
		}
	}

	return egs
}
