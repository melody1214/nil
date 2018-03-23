package s3handling

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/rpc"
	"net/url"
	"strings"
	"time"

	"math/rand"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/kv"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

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
	logger = mlog.GetLogger().WithField("package", "gw/s3handling")

	return &Handler{
		clusterMap: cmap.New(),
		authCache:  kv.New(),
	}
}

// RegisteredTo : s3 handler is registered to the multiplexer.
func (h *Handler) RegisteredTo(router *mux.Router) {
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
	// Extract credential along with user authentication.
	cred, s3Err := h.authRequest(r)
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
		AccessKey:  cred.AccessKey,
		Region:     cred.Region,
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
	// Extract credential along with user authentication.
	// cred, s3Err := h.authRequest(r)
	// if s3Err != s3.ErrNone {
	// 	s3.SendError(w, s3Err, r.RequestURI, "")
	// 	return
	// }
	// _ = cred

	// Extract bucket name and object name.
	// ex) /bucketname/object1
	// ->  bucketname/object1
	// ->  bucketName: bucketname
	// ->  objectName: object1
	bucketAndObject := strings.SplitN(strings.Trim(r.RequestURI, "/"), "/", 2)
	if len(bucketAndObject) < 2 {
		s3.SendError(w, s3.ErrInvalidURI, r.RequestURI, "")
		return
	}

	// Test code
	c := h.clusterMap.SearchCall()
	node, err := c.Type(cmap.DS).Do()
	if err != nil {
		s3.SendError(w, s3.ErrInternalError, r.RequestURI, "")
		return
	}

	rpURL, err := url.Parse("https://" + node.Addr)
	if err != nil {
		logger.WithField("method", "Handler.s3PutObject").Error(
			errors.Wrapf(
				err,
				"parse ds url failed, ds ID: %s, ds url: %s",
				node.ID.String(),
				node.Addr,
			),
		)
		s3.SendError(w, s3.ErrInternalError, r.RequestURI, "")
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(rpURL)
	r.Header.Add("Volume-Id", randomVolumeID())
	proxy.ErrorLog = log.New(logger.Writer(), "http reverse proxy", log.Lshortfile)
	proxy.ServeHTTP(w, r)
}

// For Testing
const volumes = "123456789"

func randomVolumeID() string {
	return string(volumes[rand.Intn(len(volumes))])
}

func (h *Handler) s3GetObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (h *Handler) s3DeleteObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *Handler) updateClusterMap() {
	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		logger.WithField("method", "Handler.updateClusterMap").Error(
			errors.Wrap(
				err,
				"get the latest cluster map from remote failed",
			),
		)
		return
	}

	h.clusterMap = m
}
