package cmap

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

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

// UpdateNodeStatus updates the node status of the given node ID.
func (s *Service) UpdateNodeStatus(nID ID, stat NodeStatus) error {
	return nil
}

// UpdateVolume updates the volume status of the given volume ID.
func (s *Service) UpdateVolume(volume Volume) error {
	node, err := s.manager.SearchCallNode().ID(volume.Node).Do()
	if err != nil {
		return fmt.Errorf("no such node: %v", err)
	}
	if node.Name != s.cfg.Name {
		return fmt.Errorf("only can update volumes which this node has")
	}

	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	cm := s.manager.latestCMap()
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
	vol, err := s.manager.SearchCallVolume().ID(eg.Vols[len(eg.Vols)-1]).Do()
	if err != nil {
		return fmt.Errorf("no such volume: %v", err)
	}
	node, err := s.manager.SearchCallNode().ID(vol.Node).Do()
	if node.Name != s.cfg.Name {
		return fmt.Errorf("only can update eg which this the leader volume")
	}

	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	cm := s.manager.latestCMap()
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
	return *s.manager.LatestCMap()
}

// GetLatestCMapVersion returns the latest version of cluster map.
func (s *Service) GetLatestCMapVersion() Version {
	return s.manager.latest
}

// UpdateCMap updates the new cmap manager with the given cmap.
func (s *Service) UpdateCMap(cmap *CMap) error {
	s.manager.mergeCMap(cmap)
	return nil
}

// SearchCallNode returns a new search call for finding node.
func (s *Service) SearchCallNode() *SearchCallNode {
	return s.manager.SearchCallNode()
}

// SearchCallVolume returns a new search call for finding volume.
func (s *Service) SearchCallVolume() *SearchCallVolume {
	return s.manager.SearchCallVolume()
}

// SearchCallEncGrp returns a new search call for finding encoding group.
func (s *Service) SearchCallEncGrp() *SearchCallEncGrp {
	return s.manager.SearchCallEncGrp()
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

// FindEncodingGroupByLeader finds encoding groups owned by given leader node.
func (s *Service) FindEncodingGroupByLeader(leaderNode ID) []EncodingGroup {
	m := s.manager.LatestCMap()

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
