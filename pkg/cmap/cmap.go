package cmap

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetLatest get the latest cluster map.
func GetLatest(mdsAddrs ...string) (*CMap, error) {
	Lock()
	defer Unlock()

	if len(mdsAddrs) == 0 {
		mdsAddr, err := getMdsAddr()
		if err != nil {
			return nil, fmt.Errorf("TODO: read mds address from file")
		}

		mdsAddrs = append(mdsAddrs, mdsAddr)
	}

	for _, mdsAddr := range mdsAddrs {
		m, err := getLatest(mdsAddr)
		if err != nil {
			continue
		}

		path := filePath(m.Version)
		if err = createFile(path); err != nil {
			continue
		}
		if err = encode(*m, path); err != nil {
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

func getMdsAddr() (string, error) {
	return "", fmt.Errorf("not implemented")
}

// CMap is a cluster map which includes the information about nodes.
type CMap struct {
	Version int64  `xml:"version"`
	Nodes   []Node `xml:"node"`
}

// SearchCall returns a new search call.
func (m *CMap) SearchCall() *SearchCall {
	return &SearchCall{m: m}
}

// HumanReadable returns a human readable map of the cluster.
func (m *CMap) HumanReadable() string {
	out := ""

	// Make human readable sentences for each nodes.
	for _, n := range m.Nodes {
		row := fmt.Sprintf(
			"| %4s | %s | %7s |\n",
			n.Type.String(),
			n.Addr,
			n.Stat.String(),
		)

		out += row
	}

	return out
}
