package s3

import (
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/gorilla/mux"
)

const s3PathPrefix = "/"

type s3APIHandlers struct {
	cfg *config.Gw
}

// RegisterS3APIRouter registers s3 handler to the given mux.
func RegisterS3APIRouter(cfg *config.Gw, m *mux.Router) error {
	if m == nil {
		return ErrNilMux
	}

	h := &s3APIHandlers{
		cfg: cfg,
	}

	m.PathPrefix(s3PathPrefix).HandlerFunc(h.catchAll)

	return nil
}
