package auth

import (
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// Handlers provides access an authentication domain.
type Handlers interface {
	GetSecretKey(accessKey string) (secretKey string, err error)
}

type handlers struct {
	cache Repository
	cMap  *cmap.CMap
}

// NewHandlers creates an authentication handlers with necessary dependencies.
func NewHandlers(repo Repository) Handlers {
	log = mlog.GetLogger().WithField("package", "gw/usecase/auth")

	return &handlers{
		cache: repo,
		cMap:  cmap.New(),
	}
}

// GetSecretKey returns a matched secret key with the given access key.
func (h *handlers) GetSecretKey(accessKey string) (secretKey string, err error) {
	sk, ok := h.cache.Find(accessKey)
	if ok {
		return sk, nil
	}

	secretKey, err = h.getSecretKeyFromRemote(accessKey)
	if err != nil {
		return
	}

	// Access to cache needs to hold mutex.
	// Dealing with add cache job to goroutine.
	go h.cache.Add(accessKey, secretKey)

	return
}

func (h *handlers) getSecretKeyFromRemote(accessKey string) (secretKey string, err error) {
	ctxLogger := log.WithField("method", "handlers.getSecretKeyFromRemote")

	// 1. Lookup mds from cluster map.
	mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "failed to find alive mds"))
			return "", ErrInternal
		}
	}

	// 2. Try dial to mds.
	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to dial to mds"))
		return "", ErrInternal
	}
	defer conn.Close()

	req := &nilrpc.GetCredentialRequest{AccessKey: accessKey}
	res := &nilrpc.GetCredentialResponse{}

	// 3. Request the secret key.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.GetCredential.String(), req, res); err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to call mds rpc client"))
		return "", ErrInternal
	}

	// 4. No matched key.
	if res.Exist == false {
		return "", ErrNoSuchKey
	}

	return res.SecretKey, nil
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := log.WithField("method", "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to get the latest cluster map from remote"))
		return
	}

	h.cMap = m
}
