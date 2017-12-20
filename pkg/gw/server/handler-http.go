package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func (s *Server) registerHTTPHandler() {
	s.httpMux.NewRoute().HeadersRegexp("Content-type", "application/raft").HandlerFunc(s.handleHTTP)
}

func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	rpURL, err := url.Parse("https://" + s.cfg.FirstMds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: cache proxy.
	prx := httputil.NewSingleHostReverseProxy(rpURL)
	prx.ServeHTTP(w, r)

	// DEBUG
	fmt.Printf("%#v\n", r)
}
