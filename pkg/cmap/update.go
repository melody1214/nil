package cmap

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetLatest get the latest cluster map.
func GetLatest(opts ...Option) (*CMap, error) {
	// 1. Threads can't access map file when updating.
	lock()
	defer unlock()

	// 2. Set the options.
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	// 3. If it wants the map from local, then find it and return.
	if o.fromRemote == false {
		return getLatestMapFromLocal()
	}

	// 4. If no mds address is given, then get mds address
	// from the old map file.
	mdsAddr, err := getMdsAddr()
	if err != nil {
		return nil, err
	}

	// 5. Get the latest map from the mds.
	m, err := getLatestMapFromRemote(mdsAddr)
	if err != nil {
		return nil, err
	}

	// 6. Save the map to the local.
	if err := m.Save(); err != nil {
		return nil, err
	}

	return m, nil
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
	if err := cli.Call(nilrpc.GetClusterMap.String(), req, res); err != nil {
		return nil, err
	}

	m := &CMap{
		Version: res.Version,
	}
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

func store(m *CMap) error {
	// 1. Get store file path.
	path := filePath(m.Version)

	// 2. Create empty file with the version.
	if err := createFile(path); err != nil {
		return err
	}

	// 3. Encode map data into the created file.
	if err := encode(*m, path); err != nil {
		removeFile(path)
		return err
	}

	return nil
}

func getMdsAddr() (string, error) {
	path, err := getLatestMapFile()
	if err != nil {
		return "", err
	}

	m, err := decode(path)
	if err != nil {
		return "", err
	}

	mds, err := m.SearchCall().Type(MDS).Status(Alive).Do()
	if err != nil {
		return "", err
	}
	return mds.Addr, nil
}

// Option allows to override updating parameters.
type Option func(*options)

type options struct {
	fromRemote bool
}

var defaultOptions = options{
	fromRemote: false,
}

// WithFromRemote set to get the latest cluster map from the remote address.
// The address should be available metadata server.
func WithFromRemote(enabled bool) Option {
	return func(o *options) {
		o.fromRemote = enabled
	}
}
