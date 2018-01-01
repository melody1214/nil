package server

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

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

func (s *Server) s3MakeBucket(w http.ResponseWriter, r *http.Request) {
	// Sample response.
	type SampleResponse struct {
		Code      string
		Error     string
		Message   string
		RequestId string
		Resource  string
	}

	res := SampleResponse{
		Code:    "InvalidAccessKeyId",
		Message: "The AWS access key Id you provided does not exist in our records.",
	}
	var encodedRes bytes.Buffer
	encodedRes.WriteString(xml.Header)
	e := xml.NewEncoder(&encodedRes)
	e.Encode(res)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusForbidden)
	w.Write(encodedRes.Bytes())
	w.(http.Flusher).Flush()

	// fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3RemoveBucket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3PutObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3GetObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (s *Server) s3DeleteObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}
