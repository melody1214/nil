package object

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	store Repository
}

// NewHandlers creates a object handlers with necessary dependencies.
func NewHandlers(s Repository) delivery.ObjectHandlers {
	logger = mlog.GetPackageLogger("app/mds/usecase/object")

	return &handlers{
		store: s,
	}
}

func (h *handlers) Put(req *nilrpc.ObjectPutRequest, res *nilrpc.ObjectPutResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Put")

	q := fmt.Sprintf(
		`
		INSERT INTO object (obj_name, obj_bucket, obj_encoding_group, obj_encoding_group_volume)
		SELECT '%s', b.bk_id, '%s', '%s'
		FROM bucket b
		WHERE bk_name = '%s'
		`, req.Name, req.EncodingGroup, req.EncodingGroupVolume, req.Bucket,
	)

	r, err := h.store.Execute(q)
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	a, err := r.RowsAffected()
	if err != nil {
		ctxLogger.Error(err)
		return err
	}

	if a == 0 {
		err = fmt.Errorf("no rows are affected")
		ctxLogger.Errorf("%+v", req)
		return err
	}

	return nil
}
