package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
)

func (s *Server) initHandler() {
	m := mux.NewRouter()

	// Register raft handler.
	s.registerRaftHandler(m)
	// // Register s3 handler.
	// s.registerS3Handler(m)

	s.srv.Handler = m
}

func (s *Server) registerRaftHandler(router *mux.Router) {
	router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}).HeadersRegexp("Content-type", "application/raft").HandlerFunc(s.raftServeHTTP)
}

func (s *Server) registerS3Handler(router *mux.Router) {
	// API routers.
	apiRouter := router.PathPrefix("/").Subrouter()
	bucketRouter := apiRouter.PathPrefix("/{bucket}").Subrouter()
	objectRouter := bucketRouter.PathPrefix("/{object:.+}").Subrouter()

	// Bucket request handlers
	// makeBucket : s3cmd mb s3://BUCKET
	bucketRouter.Methods("PUT").HandlerFunc(s.s3MakeBucket)
	// removeBucket : s3cmd rb s3://BUCKET
	bucketRouter.Methods("DELETE").HandlerFunc(s.s3RemoveBucket)

	// Object request handlers
	// putObject : s3cmd put FILE [FILE...] s3://BUCKET[/PREFIX]
	objectRouter.Methods("PUT").HandlerFunc(s.s3PutObject)
	// getObject : s3cmd get s3://BUCKET/OBJECT LOCAL_FILE
	objectRouter.Methods("GET").HandlerFunc(s.s3GetObject)
	// deleteObject : s3cmd del s3://BUCKET/OBJECT
	objectRouter.Methods("DELETE").HandlerFunc(s.s3DeleteObject)
}

func (s *Server) raftServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AAA")

	mds := s.cfg.FirstMds
	// mds, err := s.mdsMap.Get()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	rpURL, err := url.Parse("https://" + mds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: cache proxy.
	prx := httputil.NewSingleHostReverseProxy(rpURL)
	prx.ServeHTTP(w, r)

	// DEBUG
	fmt.Printf("%#v\n", w)
}

func (s *Server) s3MakeBucket(w http.ResponseWriter, r *http.Request) {
	// _, err := s.mdsMap.Get()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3RemoveBucket(w http.ResponseWriter, r *http.Request) {
	// _, err := s.mdsMap.Get()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3PutObject(w http.ResponseWriter, r *http.Request) {
	// _, err := s.mdsMap.Get()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3GetObject(w http.ResponseWriter, r *http.Request) {
	// _, err := s.mdsMap.Get()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3DeleteObject(w http.ResponseWriter, r *http.Request) {
	// _, err := s.mdsMap.Get()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	fmt.Fprintf(w, "%v", r)
}
