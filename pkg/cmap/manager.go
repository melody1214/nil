package cmap

import (
	"sync"
)

// manager is the object for access multiple versions of cluster maps.
type manager struct {
	latest                   Version
	cMaps                    map[Version]*CMap
	notiChannels             map[Time](chan interface{})
	stateChangedNotiChannels map[Time](chan interface{})

	mu sync.RWMutex
}

func newManager() *manager {
	// Create an empty map.
	cm := &CMap{
		Version: Version(0),
		Time:    Now(),
		Nodes:   make([]Node, 0),
	}

	m := &manager{
		cMaps:                    make(map[Version]*CMap),
		notiChannels:             make(map[Time](chan interface{})),
		stateChangedNotiChannels: make(map[Time](chan interface{})),
	}
	m.cMaps[cm.Version] = cm
	m.latest = cm.Version

	return m
}

// LatestVersion returns the latest version number in cluster maps.
func (m *manager) LatestVersion() Version {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.latest
}

// LatestCMap returns the cluster map of the latest version.
func (m *manager) LatestCMap() *CMap {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cm := m.latestCMap()
	return copyCMap(cm)
}

func (m *manager) latestCMap() *CMap {
	cm := m.cMaps[m.latest]
	return cm
}

// copyCMap do deep copy the given cmap.
func copyCMap(cm *CMap) *CMap {
	copiedMap := &CMap{
		Version: cm.Version,
		Time:    cm.Time,
	}

	// Copy nodes.
	copiedMap.Nodes = make([]Node, len(cm.Nodes))
	copy(copiedMap.Nodes, cm.Nodes)

	return copiedMap
}

// Update updates the latest cluster map.
func (m *manager) Update(cm *CMap) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.update(cm)
}

func (m *manager) update(cm *CMap) error {
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
func (m *manager) GetUpdatedNoti(ver Version) <-chan interface{} {
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
func (m *manager) GetStateChangedNoti() <-chan interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make buffered channel is important because not to be blocked
	// while in the send noti progress if the receiver had been timeout.
	notiC := make(chan interface{}, 2)
	m.stateChangedNotiChannels[Now()] = notiC

	return notiC
}

// sendStateChangedNotiToAll sends notifications to all observers when some
// states are changed.
func (m *manager) sendStateChangedNotiToAll() {
	for key, ch := range m.stateChangedNotiChannels {
		ch <- nil
		close(ch)
		delete(m.stateChangedNotiChannels, key)
	}
}

func (m *manager) sendUpdateNotiToAll() {
	for i, ch := range m.notiChannels {
		ch <- nil
		close(ch)
		delete(m.notiChannels, i)
	}
}

func (m *manager) SearchCall() *SearchCall {
	return &SearchCall{
		// Use copied the latest cluster map.
		cmap: m.LatestCMap(),
	}
}

func (m *manager) getMdsAddr() (NodeAddress, error) {
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

	mds, err := m.SearchCall().Node().Type(MDS).Status(NodeAlive).Do()
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
func (m *manager) mergeCMap(received *CMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stateChanged := false
	mine := m.latestCMap()
	if mine.Version < received.Version {
		// Mine is outdated.
		stateChanged = mergeCMap(mine, received)
		m.update(received)
	} else {
		// Same version or received is outdated.
		stateChanged = mergeCMap(received, mine)
	}

	if stateChanged {
		m.sendStateChangedNotiToAll()
	}
}

// mergeCMap compares each incarnation of members and merge to destination cmap.
func mergeCMap(src, dst *CMap) bool {
	stateChanged := false

	// Merge nodes.
	for _, sn := range src.Nodes {
		for i, dn := range dst.Nodes {
			if sn.ID != dn.ID {
				continue
			}

			merge := false
			switch sn.Stat {
			case NodeAlive:
				if dn.Stat == NodeAlive && dn.Incr < sn.Incr {
					merge = true
				}
				if dn.Stat == NodeSuspect && dn.Incr < sn.Incr {
					merge = true
				}
			case NodeSuspect:
				if dn.Stat == NodeAlive && dn.Incr <= sn.Incr {
					merge = true
				}
				if dn.Stat == NodeSuspect && dn.Incr < sn.Incr {
					merge = true
				}
			case NodeFaulty:
				merge = true
			}

			if merge == false {
				break
			}

			if dst.Nodes[i].Stat != sn.Stat {
				dst.Nodes[i].Stat = sn.Stat
				stateChanged = true
			}
			dst.Nodes[i].Incr = sn.Incr
		}
	}

	return stateChanged
}
