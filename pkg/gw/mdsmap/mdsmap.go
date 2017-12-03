package mdsmap

import (
	"context"
	"sync"
	"time"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// MdsMap is the map of local metadata servers.
type MdsMap struct {
	// MDS map info store.
	mdsMap map[string]*swimpb.Member

	mu sync.Mutex
}

// New creates mds map by asking information to the firstMdsAddr.
func New(firstMdsAddr string) (*MdsMap, error) {
	m := &MdsMap{}

	if err := m.refresh(firstMdsAddr); err != nil {
		return nil, errors.Wrapf(err, "failed to get local mds map from the first mds addr(%s)", firstMdsAddr)
	}

	go m.refresher(1 * time.Minute)

	return m, nil
}

// Dial dials grpc client to healthy mds.
// TODO: cache clients.
func (m *MdsMap) Dial() (*grpc.ClientConn, error) {
	if len(m.mdsMap) < 1 {
		return nil, errors.New("no connected mds")
	}

	for _, node := range m.mdsMap {
		cc, err := grpc.Dial(node.Addr+":"+node.Port, grpc.WithInsecure())
		if err != nil {
			continue
		}

		return cc, err
	}
	return nil, errors.New("failed to connect mds")
}

func (m *MdsMap) refresh(mdsAddr string) error {
	cc, err := grpc.Dial(mdsAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	cli := mdspb.NewMdsClient(cc)

	res, err := cli.GetClusterMap(context.Background(), &mdspb.GetClusterMapRequest{})
	if err != nil {
		return err
	}

	m.mu.Lock()

	m.mdsMap = make(map[string]*swimpb.Member)
	for _, node := range res.GetMemlist() {
		if node.GetType() != swimpb.MemberType_MDS {
			continue
		}
		m.mdsMap[node.GetUuid()] = node
	}

	m.mu.Unlock()

	return nil
}

func (m *MdsMap) doRefresh() error {
	var errs error

	for _, node := range m.mdsMap {
		err := m.refresh(node.Addr + ":" + node.Port)
		if err == nil {
			return nil
		}

		err = errors.Wrapf(err, "uuid: %s", node.GetUuid())
		if errs == nil {
			errs = err
		} else {
			errs = errors.Wrap(errs, err.Error())
		}
	}

	return errors.Wrap(errs, "failed to refresh mds map")
}

func (m *MdsMap) refresher(t time.Duration) {
	for {
		time.Sleep(t)

		if err := m.doRefresh(); err != nil {
			panic(err)
		}
	}
}
