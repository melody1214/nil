package membership

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/mds/store"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// TypeBytes returns rpc type bytes which is used to multiplexing.
func TypeBytes() []byte {
	return []byte{
		0x03, // rpcSwim
	}
}

// Handler has exposed methods for rpc server.
type Handler struct {
	store   *store.Store
	swimSrv *swim.Server
}

// NewHandler returns a new membership handler.
func NewHandler(store *store.Store, swimSrv *swim.Server) (*Handler, error) {
	log = mlog.GetLogger()

	if store == nil || swimSrv == nil {
		return nil, fmt.Errorf("invalid arguments")
	}

	return &Handler{
		store:   store,
		swimSrv: swimSrv,
	}, nil
}
