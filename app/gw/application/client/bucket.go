package client

import (
	"net/http"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/gw/application/auth"
	"github.com/chanyoung/nil/app/gw/domain/model/cred"
	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

// MakeBucketHandler handles the client request for making a new bucket.
func (h *handlers) MakeBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.MakeBucketHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		ctxLogger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sk, err := h.authHandlers.GetSecretKey(cred.Key(req.AccessKey()))
	if err == auth.ErrInternal {
		req.SendInternalError()
		return
	} else if err == auth.ErrNoSuchKey {
		req.SendNoSuchKey()
		return
	}

	if req.Auth(sk.String()) == false {
		req.SendIncorrectKey()
		return
	}

	if err = h.makeBucket(
		req.AccessKey(),
		req.Region(),
		req.Bucket(),
	); err != nil {
		req.SendInternalError()
		return
	}

	req.SendSuccess()
}

func (h *handlers) makeBucket(accessKey, region, bucket string) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.makeBucket")

	// // 1. Lookup mds from cmap.
	// mds, err := h.cmapAPI.SearchCall().Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
	// if err != nil {
	// 	ctxLogger.Error(err)
	// 	return errInternal
	// }

	// Dialing to mds for making rpc connection.
	// conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
	conn, err := nilrpc.Dial("localhost:51000", nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(err)
		return errInternal
	}
	defer conn.Close()

	// Fill the request and prepare response object.
	req := &nilrpc.MACMakeBucketRequest{
		AccessKey:  accessKey,
		Region:     region,
		BucketName: bucket,
	}
	res := &nilrpc.MACMakeBucketResponse{}

	// Call 'AddBucket' procedure and handling errors.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsAccountMakeBucket.String(), req, res); err != nil {
		// Not mysql error, unknown error.
		ctxLogger.Error(err)
		return errInternal
	} else if res.S3ErrCode != s3.ErrNone {
		// Kind of mysql error, mds would change it to s3.ErrorCode.
		// TODO: change to not s3 related.
		ctxLogger.Error(res.S3ErrCode)
		return errInternal
	}

	return nil
}

// RemoveBucketHandler handles the client request for removing a bucket.
func (h *handlers) RemoveBucketHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}
