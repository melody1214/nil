package mdsmap

import (
	"context"
	"sync"
	"time"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MdsMap is the map of local metadata servers.
type MdsMap struct {
	secuCfg *config.Security
	// MDS map info store.
	mdsMap map[string]*swimpb.Member

	mu sync.Mutex
}

// New creates mds map by asking information to the firstMdsAddr.
func New(secuCfg *config.Security) (*MdsMap, error) {
	return &MdsMap{
		secuCfg: secuCfg,
	}, nil
}

// Start fills mds map information into the map and start periodic refreshing.
func (m *MdsMap) Start(mdsAddr string) error {
	if err := m.refresh(mdsAddr); err != nil {
		return errors.Wrapf(err, "failed to get local mds map from the first mds addr(%s)", mdsAddr)
	}

	go m.refresher(1 * time.Minute)
	return nil
}

// Get returns random alive member of mds map.
func (m *MdsMap) Get() (addr string, err error) {
	for _, m := range m.mdsMap {
		if m.Status == swimpb.Status_ALIVE {
			return m.GetAddr() + ":" + m.GetPort(), nil
		}
	}

	return "", errors.New("empty mdsmap")
}

func (m *MdsMap) refresher(t time.Duration) {
	for {
		time.Sleep(t)

		if err := m.doRefresh(); err != nil {
			panic(err)
		}
	}
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

func (m *MdsMap) refresh(mdsAddr string) error {
	creds, err := credentials.NewClientTLSFromFile(
		m.secuCfg.CertsDir+"/"+m.secuCfg.RootCAPem,
		"localhost",
	)
	if err != nil {
		return err
	}

	cc, err := grpc.Dial(mdsAddr, grpc.WithTransportCredentials(creds))
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
