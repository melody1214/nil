package cmap

import (
	"sync"
	"time"
)

// Controller is the object for access multiple versions of cluster maps.
type Controller struct {
	latest       Version
	cMaps        map[Version]*CMap
	notiChannels map[time.Time](chan interface{})

	mu sync.RWMutex
}

// NewController creates the cluster map object with an initial cluster map
// with the given coordinator address.
func NewController(coordinator string) (*Controller, error) {
	// Create an empty map.
	cm := &CMap{
		Version: Version(0),
		Time:    time.Now().UTC().String(),
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

	// Save to local.
	err := cm.Save()
	if err != nil {
		return nil, err
	}

	c := &Controller{
		cMaps:        make(map[Version]*CMap),
		notiChannels: make(map[time.Time](chan interface{})),
	}
	c.cMaps[cm.Version] = cm
	c.latest = cm.Version

	return c, nil
}

// LatestVersion returns the latest version number in cluster maps.
func (c *Controller) LatestVersion() Version {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.latest
}

// LatestCMap returns the cluster map of the latest version.
func (c *Controller) LatestCMap() CMap {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cm := c.cMaps[c.latest]
	return *cm
}

// Update updates the latest cluster map.
func (c *Controller) Update(cm *CMap) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If version is less or equal to current latest version,
	// then we don't need to do anything.
	if cm.Version <= c.latest {
		return nil
	}

	// Save the latest map to the local.
	if err := cm.Save(); err != nil {
		return err
	}

	c.cMaps[cm.Version] = cm
	c.latest = cm.Version

	c.sendUpdateNotiToAll()

	return nil
}

// GetUpdatedNoti returns a channel which will send notification when
// the higher version of cluster map is created.
func (c *Controller) GetUpdatedNoti(ver Version) <-chan interface{} {
	// Make buffered channel is important because not to be blocked
	// while in the send noti progress if the receiver had been timeout.
	notiC := make(chan interface{}, 2)

	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		// Add notification channel.
		if c.latest <= ver {
			c.notiChannels[time.Now()] = notiC
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

func (c *Controller) sendUpdateNotiToAll() {
	for i, ch := range c.notiChannels {
		ch <- nil
		close(ch)
		delete(c.notiChannels, i)
	}
}

// SearchCall returns a new search call.
func (c *Controller) SearchCall() *SearchCall {
	return &SearchCall{
		m: c,
		i: ID(-1),
		n: "",
	}
}

func (c *Controller) getMdsAddr() (string, error) {
	var cm *CMap
	// If fail to get cluster map from the memory,
	// then look up in the file system directory.
	if cm = c.cMaps[c.latest]; cm == nil {
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

	mds, err := c.SearchCall().Type(MDS).Status(Alive).Do()
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
