package object

import (
	"fmt"

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

func (h *handlers) Get(req *nilrpc.ObjectGetRequest, res *nilrpc.ObjectGetResponse) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Get")

	q := fmt.Sprintf(
		`
		SELECT
			obj_encoding_group, obj_encoding_group_volume
		FROM
			object
		WHERE
			obj_name = '%s'
		`, req.Name,
	)

	row := h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	var v string
	err := row.Scan(&res.EncodingGroup, &v)
	if err != nil {
		return err
	}

	var col string
	switch v {
	case "first":
		col = "eg_first_volume"
	case "second":
		col = "eg_second_volume"
	case "third":
		col = "eg_third_volume"
	default:
		err = fmt.Errorf("object is existed but failed to retrieve proper volume from encoding group")
		ctxLogger.Error(err)
		return err
	}

	q = fmt.Sprintf(
		`
		SELECT
			%s
		FROM
			encoding_group
		WHERE
			eg_id = '%d'
		`, col, res.EncodingGroup,
	)

	row = h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	err = row.Scan(&res.EncodingGroupVolumeID)
	if err != nil {
		return err
	}

	q = fmt.Sprintf(
		`
		SELECT
			vl_node
		FROM
			volume
		WHERE
			vl_id = '%d'
		`, res.EncodingGroupVolumeID,
	)

	row = h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	err = row.Scan(&res.DsID)
	if err != nil {
		return err
	}

	return nil
}

// Handlers is the interface that provides object domain's rpc handlers.
type Handlers interface {
	Put(req *nilrpc.ObjectPutRequest, res *nilrpc.ObjectPutResponse) error
	Get(req *nilrpc.ObjectGetRequest, res *nilrpc.ObjectGetResponse) error
}
