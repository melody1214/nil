package s3handling

import (
	"fmt"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/kv"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// TypeBytes returns rpc type bytes which is used to multiplexing.
func TypeBytes() []byte {
	return []byte{
		0x44, // 'D' of DELETE
		0x47, // 'G' of GET
		0x50, // 'P' of POST, PUT
	}
}

// Handler offers methods that can handle s3 requests.
type Handler struct {
	clusterMap *cmap.CMap

	authCache kv.DB
}

// NewHandler returns a new rpc handler.
func NewHandler() *Handler {
	log = mlog.GetLogger()

	return &Handler{
		clusterMap: cmap.NewCMap(),
		authCache:  kv.New(),
	}
}

// Register registers s3 handling methods to the http mux.
func (h *Handler) Register(router *mux.Router) {
	// API routers.
	apiRouter := router.PathPrefix("/").Subrouter()
	bucketRouter := apiRouter.PathPrefix("/{bucket}").Subrouter()
	objectRouter := bucketRouter.PathPrefix("/{object:.+}").Subrouter()

	// Bucket request handlers
	// makeBucket : s3cmd mb s3://BUCKET
	bucketRouter.Methods("PUT").HandlerFunc(h.s3MakeBucket)
	// removeBucket : s3cmd rb s3://BUCKET
	bucketRouter.Methods("DELETE").HandlerFunc(h.s3RemoveBucket)

	// Object request handlers
	// putObject : s3cmd put FILE [FILE...] s3://BUCKET[/PREFIX]
	objectRouter.Methods("PUT").HandlerFunc(h.s3PutObject)
	// getObject : s3cmd get s3://BUCKET/OBJECT LOCAL_FILE
	objectRouter.Methods("GET").HandlerFunc(h.s3GetObject)
	// deleteObject : s3cmd del s3://BUCKET/OBJECT
	objectRouter.Methods("DELETE").HandlerFunc(h.s3DeleteObject)
}

func (h *Handler) s3MakeBucket(w http.ResponseWriter, r *http.Request) {
	// Extract access key along with user authentication.
	accessKey, s3Err := h.authRequest(r)
	if s3Err != s3.ErrNone {
		s3.SendError(w, s3Err, r.RequestURI, "")
		return
	}

	// Extract bucket name.
	// ex) /bucketname/ -> bucketname
	bucketName := strings.Trim(r.RequestURI, "/")

	// 2. Lookup mds from cluster map.
	mds, err := h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		mds, err = h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			s3.SendError(w, s3.ErrInternalError, r.RequestURI, "")
			return
		}
	}

	// Dialing to mds for making rpc connection.
	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
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

func (h *Handler) s3RemoveBucket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (h *Handler) s3PutObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (h *Handler) s3GetObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (h *Handler) s3DeleteObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *Handler) updateClusterMap() {
	m, err := cmap.GetLatest()
	if err != nil {
		log.Error(err)
		return
	}

	h.clusterMap = m
}
