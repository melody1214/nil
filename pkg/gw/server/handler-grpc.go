package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
)

func (s *Server) registerGrpcHandler() {
	s.httpMux.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.ProtoMajor == 2
	}).HeadersRegexp("Content-type", "application/grpc").HandlerFunc(s.handleGrpc)
}

func (s *Server) handleGrpc(w http.ResponseWriter, r *http.Request) {
	rpURL, err := url.Parse("https://" + s.cfg.FirstMds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prx := httputil.NewSingleHostReverseProxy(rpURL)

	// Write gRPC dafault headers.
	w.Header().Add("Trailer", "Grpc-Status")
	w.Header().Add("Trailer", "Grpc-Message")
	w.Header().Add("Trailer", "Grpc-Status-Details-Bin")
	prx.ServeHTTP(w, r)

	// DEBUG
	fmt.Printf("%+v\n", w)
}
