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
	cMap                *cmap.Controller
}

// NewHandlers creates a client handlers with necessary dependencies.
func NewHandlers(cMap *cmap.Controller, f *cr.RequestEventFactory, s Repository) delivery.ObjectHandlers {
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
