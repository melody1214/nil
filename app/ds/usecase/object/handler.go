package object

import (
	"github.com/chanyoung/nil/app/ds/delivery"
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
	cMap                *cmap.CMap
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(f *cr.RequestEventFactory, s Repository) delivery.ObjectHandlers {
	logger = mlog.GetPackageLogger("app/ds/usecase/object")

	enc := newEncoder(s)
	go enc.Run()

	return &handlers{
		requestEventFactory: f,
		encoder:             enc,
		store:               s,
		cMap:                cmap.New(),
	}
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	h.cMap = m
}
