package raft

import (
	"net/http"

	"github.com/chanyoung/nil/pkg/gw/mdsmap"
	"github.com/chanyoung/nil/pkg/raft/raftpb"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

type raftHandlers struct {
	cfg    *config.Gw
	mdsMap *mdsmap.MdsMap
	g      *grpc.Server
}

// RegisterRaftRouter registers raft handler to the given mux.
func RegisterRaftRouter(cfg *config.Gw, m *mux.Router) error {
	if m == nil {
		return ErrNilMux
	}

	mdsMap, err := mdsmap.New(cfg.FirstMds)
	if err != nil {
		return err
	}

	h := &raftHandlers{
		cfg:    cfg,
		mdsMap: mdsMap,
		g:      grpc.NewServer(),
	}

	raftpb.RegisterRaftServer(h.g, h)

	m.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.ProtoMajor == 2
	}).HeadersRegexp("Content-Type", "application/grpc").HandlerFunc(h.g.ServeHTTP)

	return nil
}
