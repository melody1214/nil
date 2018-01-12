package server

import (
	"fmt"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
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
	accessKey, s3Err := s.authRequest(r)
	if s3Err != s3.ErrNone {
		s3.SendError(w, s3Err, r.RequestURI, "")
	}

	bucketName := strings.Trim(r.RequestURI, "/")

	conn, err := nilrpc.Dial(s.cfg.FirstMds, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer conn.Close()

	req := &nilrpc.AddBucketRequest{
		AccessKey:  accessKey,
		BucketName: bucketName,
	}
	res := &nilrpc.AddBucketResponse{}
	fmt.Printf("%v\n", req)

	cli := rpc.NewClient(conn)
	if err := cli.Call("Server.AddBucket", req, res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
