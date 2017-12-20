package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) registerHTTPHandler() {
	s.httpMux.NewRoute().HeadersRegexp("Content-type", "application/raft").HandlerFunc(s.handleJoin)
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	// DEBUG
	fmt.Printf("%+v\n", r)

	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}