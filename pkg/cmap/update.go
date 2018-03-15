package cmap

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetLatest get the latest cluster map.
func GetLatest(mdsAddrs ...string) (*CMap, error) {
	// 1. Threads can't access map file when updating.
	lock()
	defer unlock()

	// 2. If no mds address is given, then get mds address
	// from the old map file.
	if len(mdsAddrs) == 0 {
		mdsAddr, err := getMdsAddr()
		if err != nil {
			return nil, err
		}

		mdsAddrs = append(mdsAddrs, mdsAddr)
	}

	// 3. Get the latest map from the mds.
	for _, mdsAddr := range mdsAddrs {
		m, err := getLatest(mdsAddr)
		if err != nil {
			continue
		}

		if err := store(m); err != nil {
			continue
		}

		return m, nil
	}

	return nil, fmt.Errorf("couldn't get the mds address")
}

func getLatest(mdsAddr string) (*CMap, error) {
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
