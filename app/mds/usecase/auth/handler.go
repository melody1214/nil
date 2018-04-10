package auth

import (
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store Repository
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/auth")

	return &handlers{
		store: s,
	}
}

// GetCredential returns matching secret key with the given access key.
func (h *handlers) GetCredential(req *nilrpc.MAUGetCredentialRequest, res *nilrpc.MAUGetCredentialResponse) error {
	res.AccessKey = req.AccessKey

	sk, err := h.store.FindSecretKey(req.AccessKey)
	if err == nil {
		res.Exist = true
		res.SecretKey = sk
	} else if err == repository.ErrNotExist {
		res.Exist = false
	} else {
		return err
	}

	return nil
}

// Handlers is the interface that provides auth domain's rpc handlers.
type Handlers interface {
	GetCredential(req *nilrpc.MAUGetCredentialRequest, res *nilrpc.MAUGetCredentialResponse) error
}
