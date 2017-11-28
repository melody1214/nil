package grpc

import (
	"net/http"

	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/gorilla/mux"
)

type gRPCHandlers struct {
	cfg *config.Gw
}

// RegisterGRPCRouter registers gRPC handler to the given mux.
func RegisterGRPCRouter(cfg *config.Gw, m *mux.Router) error {
	if m == nil {
		return ErrNilMux
	}

	h := &gRPCHandlers{
		cfg: cfg,
	}

	m.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.ProtoMajor == 2
	}).HeadersRegexp("Content-Type", "application/grpc").HandlerFunc(h.test)

	return nil
}
