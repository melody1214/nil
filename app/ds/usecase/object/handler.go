package object

import (
	"net/http"

	cr "github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	requestEventFactory *cr.RequestEventFactory
	store               Repository
	encoder             *encoder
	cMap                *cmap.Controller
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cMap *cmap.Controller, f *cr.RequestEventFactory, s Repository) ObjectHandlers {
	logger = mlog.GetPackageLogger("app/ds/usecase/object")

	enc := newEncoder(cMap, s)
	go enc.Run()

	return &handlers{
		requestEventFactory: f,
		encoder:             enc,
		store:               s,
		cMap:                cMap,
	}
}

// ObjectHandlers is the interface that provides client http handlers.
type ObjectHandlers interface {
	PutObjectHandler(w http.ResponseWriter, r *http.Request)
	GetObjectHandler(w http.ResponseWriter, r *http.Request)
	DeleteObjectHandler(w http.ResponseWriter, r *http.Request)
}
