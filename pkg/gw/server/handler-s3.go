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
	// Extract access key along with user authentication.
	accessKey, s3Err := s.authRequest(r)
	if s3Err != s3.ErrNone {
		s3.SendError(w, s3Err, r.RequestURI, "")
	}

	// Extract bucket name.
	// ex) /bucketname/ -> bucketname
	bucketName := strings.Trim(r.RequestURI, "/")

	// Dialing to mds for making rpc connection.
	conn, err := nilrpc.Dial(s.cfg.FirstMds, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		s3.SendError(w, s3.ErrInternalError, r.RequestURI, "")
		return
	}
	defer conn.Close()

	// Fill the request and prepare response object.
	req := &nilrpc.AddBucketRequest{
		AccessKey:  accessKey,
		BucketName: bucketName,
	}
	res := &nilrpc.AddBucketResponse{}

	// Call 'AddBucket' procedure and handling errors.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.AddBucket.String(), req, res); err != nil {
		// Not mysql error, unknown error.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if res.S3ErrCode != s3.ErrNone {
		// Kind of mysql error, mds would change it to s3.ErrorCode.
		s3.SendError(w, res.S3ErrCode, r.RequestURI, "")
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
