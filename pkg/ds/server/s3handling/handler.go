package s3handling

import (
	"fmt"
	"net/http"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/ds/store"
	"github.com/chanyoung/nil/pkg/ds/store/request"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

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
	store      store.Service
}

// New returns a new s3 handler.
func New(s store.Service) (*Handler, error) {
	logger = mlog.GetLogger()

	if s == nil {
		return nil, fmt.Errorf("nil store object")
	}

	return &Handler{
		store:      s,
		clusterMap: cmap.New(),
	}, nil
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
	fmt.Fprintf(w, "%v", r)
}

func (h *Handler) s3RemoveBucket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

func (h *Handler) s3PutObject(w http.ResponseWriter, r *http.Request) {
	req := &request.Request{
		Op:  request.Write,
		Vol: r.Header.Get("Volume-Id"),
		Oid: "AAA",

		In: r.Body,
	}

	logger.Infof("%+v", req)

	err := h.store.Push(req)
	if err != nil {
		logger.Error(err)
		s3.SendError(w, s3.ErrInternalError, r.RequestURI, "")
		return
	}

	err = req.Wait()
	if err != nil {
		logger.Error(err)
		s3.SendError(w, s3.ErrInternalError, r.RequestURI, "")
	}
	logger.Info("Finished")

	s3.SendSuccess(w)
	// fmt.Fprintf(w, "%v", r)
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
		logger.Error(err)
		return
	}

	h.clusterMap = m
}
