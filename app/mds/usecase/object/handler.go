package object

import (
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store Repository
}

// NewHandlers creates a object handlers with necessary dependencies.
func NewHandlers(s Repository) Handlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/object")

	return &handlers{
		store: s,
	}
}

// ObjInfo holds the information which is required to access the object.
type ObjInfo struct {
	Name   string
	Bucket string
	EncGrp cmap.ID
	Vol    cmap.ID
	Node   cmap.ID
}

func (h *handlers) Put(req *nilrpc.MOBObjectPutRequest, res *nilrpc.MOBObjectPutResponse) error {
	return h.store.Put(&ObjInfo{
		Name:   req.Name,
		Bucket: req.Bucket,
		EncGrp: req.EncodingGroup,
		Vol:    req.Volume,
	})
}

func (h *handlers) Get(req *nilrpc.MOBObjectGetRequest, res *nilrpc.MOBObjectGetResponse) error {
	o, err := h.store.Get(req.Name)
	if err != nil {
		return nil
	}

	res.EncodingGroupID = o.EncGrp
	res.VolumeID = o.Vol
	res.DsID = o.Node

	return nil
}

// Handlers is the interface that provides object domain's rpc handlers.
type Handlers interface {
	Put(req *nilrpc.MOBObjectPutRequest, res *nilrpc.MOBObjectPutResponse) error
	Get(req *nilrpc.MOBObjectGetRequest, res *nilrpc.MOBObjectGetResponse) error
}
