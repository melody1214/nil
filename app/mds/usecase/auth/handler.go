package auth

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store Repository
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(s Repository) delivery.AuthHandlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/auth")

	return &handlers{
		store: s,
	}
}

// TODO: CQRS

// GetCredential returns matching secret key with the given access key.
func (h *handlers) GetCredential(req *nilrpc.GetCredentialRequest, res *nilrpc.GetCredentialResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
			secret_key
		FROM
			user
		WHERE
			access_key = '%s'
		`, req.AccessKey,
	)

	res.AccessKey = req.AccessKey
	err := h.store.QueryRow(q).Scan(&res.SecretKey)
	if err == nil {
		res.Exist = true
	} else if err == sql.ErrNoRows {
		res.Exist = false
	} else {
		return err
	}

	return nil
}
