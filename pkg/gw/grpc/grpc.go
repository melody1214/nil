package grpc

import (
	"net/http"

	"github.com/chanyoung/nil/pkg/gw/mdsmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/gorilla/mux"
)

type gRPCHandlers struct {
	cfg    *config.Gw
	mdsMap *mdsmap.MdsMap
}

// RegisterGRPCRouter registers gRPC handler to the given mux.
func RegisterGRPCRouter(cfg *config.Gw, m *mux.Router) error {
	if m == nil {
		return ErrNilMux
	}

	mdsMap, err := mdsmap.New(cfg.FirstMds)
	if err != nil {
		return err
	}

	h := &gRPCHandlers{
		cfg:    cfg,
		mdsMap: mdsMap,
	}

	m.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.ProtoMajor == 2
	}).HeadersRegexp("Content-Type", "application/grpc").HandlerFunc(h.test)

	return nil
}
