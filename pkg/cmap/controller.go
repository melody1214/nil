package cmap

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
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
		Nodes:   make([]Node, 1),
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
func (c *Controller) Update(opts ...Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	var cm *CMap
	if o.withFile {
		cm = o.file
	} else {
		// Get mds address from the map file.
		mdsAddr, err := c.getMdsAddr()
		if err != nil {
			return err
		}

		// Get the latest map from the mds.
		cm, err = getLatestMapFromRemote(mdsAddr)
		if err != nil {
			return err
		}
	}

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

// Option allows to override updating parameters.
type Option func(*options)

type options struct {
	withFile bool
	file     *CMap
}

var defaultOptions = options{
	withFile: false,
	file:     nil,
}

// WithFile do update with the given cluster map file.
func WithFile(m *CMap) Option {
	return func(o *options) {
		o.withFile = true
		o.file = m
	}
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

func getLatestMapFromRemote(mdsAddr string) (*CMap, error) {
	conn, err := nilrpc.Dial(mdsAddr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &nilrpc.GetClusterMapRequest{}
	res := &nilrpc.GetClusterMapResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClustermapGetClusterMap.String(), req, res); err != nil {
		return nil, err
	}

	m := &CMap{Version: Version(res.Version)}
	for _, n := range res.Nodes {
		m.Nodes = append(m.Nodes, Node{
			ID:   ID(n.ID),
			Name: n.Name,
			Addr: n.Addr,
			Type: Type(n.Type),
			Stat: Status(n.Stat),
		})
	}

	return m, nil
}
