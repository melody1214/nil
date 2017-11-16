package s3

import (
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/gorilla/mux"
)

type s3APIHandlers struct {
	cfg *config.Gw
}

// RegisterS3APIRouter registers s3 handler to the given mux.
func RegisterS3APIRouter(cfg *config.Gw, m *mux.Router) error {
	if m == nil {
		return ErrNilMux
	}

	h := &s3APIHandlers{
		cfg: cfg,
	}

	// API routers.
	apiRouter := m.PathPrefix("/").Subrouter()
	bucketRouter := apiRouter.PathPrefix("/{bucket}").Subrouter()
	objectRouter := bucketRouter.PathPrefix("/{object:.+}").Subrouter()

	// Bucket request handlers
	// makeBucket : s3cmd mb s3://BUCKET
	bucketRouter.Methods("PUT").HandlerFunc(h.makeBucket)
	// removeBucket : s3cmd rb s3://BUCKET
	bucketRouter.Methods("DELETE").HandlerFunc(h.removeBucket)

	// Object request handlers
	// putObject : s3cmd put FILE [FILE...] s3://BUCKET[/PREFIX]
	objectRouter.Methods("PUT").HandlerFunc(h.putObject)
	// getObject : s3cmd get s3://BUCKET/OBJECT LOCAL_FILE
	objectRouter.Methods("GET").HandlerFunc(h.getObject)
	// deleteObject : s3cmd del s3://BUCKET/OBJECT
	objectRouter.Methods("DELETE").HandlerFunc(h.deleteObject)

	return nil
}
