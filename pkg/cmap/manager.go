package cmap

import (
	"math/rand"
	"sync"
	"time"
)

// cMapManager is the object for access multiple versions of cluster maps.
type cMapManager struct {
	latest                   Version
	cMaps                    map[Version]*CMap
	notiChannels             map[Time](chan interface{})
	stateChangedNotiChannels map[Time](chan interface{})
	random                   *rand.Rand

	mu sync.RWMutex
}

// newCMapManager creates the cluster map manager with an initial cluster map
// with the given coordinator address.
func newCMapManager(coordinator NodeAddress) (*cMapManager, error) {
	// Create an empty map.
	cm := &CMap{
		Version: Version(0),
		Time:    Now(),
		Nodes:   make([]Node, 1),
		Vols:    make([]Volume, 0),
		EncGrps: make([]EncodingGroup, 0),
	}

	// Set the mds.
	cm.Nodes[0] = Node{
		Addr: coordinator,
		Type: MDS,
		Stat: Alive,
	}

	// Save to file system.
	err := cm.Save()
	if err != nil {
		return nil, err
	}

	m := &cMapManager{
		cMaps:                    make(map[Version]*CMap),
		notiChannels:             make(map[Time](chan interface{})),
		stateChangedNotiChannels: make(map[Time](chan interface{})),
		random: rand.New(rand.NewSource(time.Now().Unix())),
	}
	m.cMaps[cm.Version] = cm
	m.latest = cm.Version

	return m, nil
}

// LatestVersion returns the latest version number in cluster maps.
func (m *cMapManager) LatestVersion() Version {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.latest
}

// LatestCMap returns the cluster map of the latest version.
func (m *cMapManager) LatestCMap() CMap {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cm := m.latestCMap()
	return copyCMap(cm)
}

func (m *cMapManager) latestCMap() *CMap {
	cm := m.cMaps[m.latest]
	return cm
}

// copyCMap do deep copy the given cmap.
func copyCMap(cm *CMap) CMap {
	copiedMap := CMap{
		Version: cm.Version,
		Time:    cm.Time,
	}

	// Copy nodes.
	copiedMap.Nodes = make([]Node, len(cm.Nodes))
	copy(copiedMap.Nodes, cm.Nodes)
	for i, n := range cm.Nodes {
		copiedMap.Nodes[i].Vols = make([]ID, len(n.Vols))
		copy(copiedMap.Nodes[i].Vols, n.Vols)
	}

	// Copy volumes.
	copiedMap.Vols = make([]Volume, len(cm.Vols))
	copy(copiedMap.Vols, cm.Vols)
	for i, v := range cm.Vols {
		copiedMap.Vols[i].EncGrps = make([]ID, len(v.EncGrps))
		copy(copiedMap.Vols[i].EncGrps, v.EncGrps)
	}

	// Copy encoding groups.
	copiedMap.EncGrps = make([]EncodingGroup, len(cm.EncGrps))
	copy(copiedMap.EncGrps, cm.EncGrps)
	for i, eg := range cm.EncGrps {
		copiedMap.EncGrps[i].Vols = make([]ID, len(eg.Vols))
		copy(copiedMap.EncGrps[i].Vols, eg.Vols)
	}

	return copiedMap
}

// Update updates the latest cluster map.
func (m *cMapManager) Update(cm *CMap) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.update(cm)
}

func (m *cMapManager) update(cm *CMap) error {
	// If version is less or equal to current latest version,
	// then we don't need to do anything.
	if cm.Version <= m.latest {
		return nil
	}

	// Save the latest map to the local.
	if err := cm.Save(); err != nil {
		return err
	}

	m.cMaps[cm.Version] = cm
	m.latest = cm.Version

	m.sendUpdateNotiToAll()

	return nil
}

// GetUpdatedNoti returns a channel which will send notification when
// the higher version of cluster map is created.
func (m *cMapManager) GetUpdatedNoti(ver Version) <-chan interface{} {
	// Make buffered channel is important because not to be blocked
	// while in the send noti progress if the receiver had been timeout.
	notiC := make(chan interface{}, 2)

	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		// Add notification channel.
		if m.latest <= ver {
			m.notiChannels[Now()] = notiC
			return
		}

		// Already have the higher version of cluster map.
		// Send notification and close the channel.
		notiC <- nil
		close(notiC)
		return
	}()

	return notiC
}

// GetStateChangedNoti returns a channel which will send notification when
// some cluster map member's state are changed.
func (m *cMapManager) GetStateChangedNoti() <-chan interface{} {
	// Make buffered channel is important because not to be blocked
	// while in the send noti progress if the receiver had been timeout.
	notiC := make(chan interface{}, 2)

	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		m.stateChangedNotiChannels[Now()] = notiC
	}()

	return notiC
}

// StateChanged means some states are changed and sends notifications to all observers.
func (m *cMapManager) StateChanged() {
	for i, ch := range m.stateChangedNotiChannels {
		ch <- nil
		close(ch)
		delete(m.stateChangedNotiChannels, i)
	}
}

func (m *cMapManager) sendUpdateNotiToAll() {
	for i, ch := range m.notiChannels {
		ch <- nil
		close(ch)
		delete(m.notiChannels, i)
	}
}

// SearchCallNode returns a new search call for finding node.
func (m *cMapManager) SearchCallNode() *SearchCallNode {
	return &SearchCallNode{
		manager: m,
		id:      ID(-1),
		name:    NodeName(""),
		nType:   NodeType(-1),
		status:  NodeStatus(unknown),
	}
}

// SearchCallEncGrp returns a new search call for finding encoding group.
func (m *cMapManager) SearchCallEncGrp() *SearchCallEncGrp {
	return &SearchCallEncGrp{
		manager: m,
		id:      ID(-1),
		status:  EncodingGroupStatus(unknown),
		random:  false,
	}
}

// SearchCallVolume returns a new search call for finding volume.
func (m *cMapManager) SearchCallVolume() *SearchCallVolume {
	return &SearchCallVolume{
		manager: m,
		id:      ID(-1),
		status:  VolumeStatus(unknown),
	}
}

func (m *cMapManager) getMdsAddr() (NodeAddress, error) {
	var cm *CMap
	// If fail to get cluster map from the memory,
	// then look up in the file system directory.
	if cm = m.cMaps[m.latest]; cm == nil {
		path, err := getLatestMapFile()
		if err != nil {
			return "", err
		}

		m, err := decode(path)
		if err != nil {
			return "", err
		}

		cm = &m
	}

	mds, err := m.SearchCallNode().Type(MDS).Status(Alive).Do()
	if err != nil {
		return "", err
	}
	return mds.Addr, nil
}

func getLatestMapFromLocal() (*CMap, error) {
	// 1. Get the latest map full path from the local.
	path, err := getLatestMapFile()
	if err != nil {
		return nil, err
	}

	// 2. Decode the file and return the cluster map.
	m, err := decode(path)
	return &m, err
}

// mergeCMap compares cluster map mine with received by swim protocol one,
// and merge it to newer version. Set merged cmap to the cluster map manager.
func (m *cMapManager) mergeCMap(received *CMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mine := m.latestCMap()
	if mine.Version < received.Version {
		// Mine is outdated.
		mergeCMap(mine, received)
		m.update(received)
	} else {
		// Same version or received is outdated.
		mergeCMap(received, mine)
	}
}

// mergeCMap compares each incarnation of members and merge to destination cmap.
func mergeCMap(src, dst *CMap) {
	// Merge nodes.
	for _, sn := range src.Nodes {
		for i, dn := range dst.Nodes {
			if sn.ID != dn.ID {
				continue
			}

			merge := false
			switch sn.Stat {
			case Alive:
				if dn.Stat == Alive && dn.Incr < sn.Incr {
					merge = true
				}
				if dn.Stat == Suspect && dn.Incr < sn.Incr {
					merge = true
				}
			case Suspect:
				if dn.Stat == Alive && dn.Incr <= sn.Incr {
					merge = true
				}
				if dn.Stat == Suspect && dn.Incr < sn.Incr {
					merge = true
				}
			case Faulty:
				merge = true
			}

			if merge {
				dst.Nodes[i].Stat = sn.Stat
				dst.Nodes[i].Incr = sn.Incr
			}
		}
	}

	// Merge volumes.
	for _, sv := range src.Vols {
		for i, dv := range dst.Vols {
			if sv.ID != dv.ID {
				continue
			}

			if sv.Incr < dv.Incr {
				break
			}

			if sv.Incr == dv.Incr && dv.Stat == Failed {
				break
			}

			dst.Vols[i].Size = sv.Size
			dst.Vols[i].Stat = sv.Stat
			dst.Vols[i].Incr = sv.Incr
		}
	}

	// Merge encoding groups.
	for _, se := range src.EncGrps {
		for i, de := range dst.EncGrps {
			if se.ID != de.ID {
				continue
			}

			if se.Incr < de.Incr {
				break
			}

			if se.Incr == de.Incr && de.Stat == EGFaulty {
				break
			}

			dst.EncGrps[i].Size = se.Size
			dst.EncGrps[i].Used = se.Used
			dst.EncGrps[i].Free = se.Free
			dst.EncGrps[i].Stat = se.Stat
			dst.EncGrps[i].Incr = se.Incr
		}
	}
}
