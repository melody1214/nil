package client

import (
	"net/http"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/gw/usecase/auth"
	"github.com/chanyoung/nil/pkg/client"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
)

func (h *handlers) MakeBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctxLogger := log.WithField("method", "handlers.MakeBucketHandler")

	req, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		ctxLogger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	sk, err := h.authHandlers.GetSecretKey(req.AccessKey())
	if err == auth.ErrInternal {
		req.SendInternalError()
	} else if err == auth.ErrNoSuchKey {
		req.SendNoSuchKey()
	}

	if req.Auth(sk) == false {
		req.SendIncorrectKey()
	}

	if err = h.makeBucket(
		req.AccessKey(),
		req.Region(),
		req.Bucket(),
	); err != nil {
		req.SendInternalError()
	}
}

func (h *handlers) makeBucket(accessKey, region, bucket string) error {
	ctxLogger := log.WithField("method", "handlers.makeBucket")

	// 1. Lookup mds from cluster map.
	mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			ctxLogger.Error(err)
			return errInternal
		}
	}

	// Dialing to mds for making rpc connection.
	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(err)
		return errInternal
	}
	defer conn.Close()

	// Fill the request and prepare response object.
	req := &nilrpc.AddBucketRequest{
		AccessKey:  accessKey,
		Region:     region,
		BucketName: bucket,
	}
	res := &nilrpc.AddBucketResponse{}

	// Call 'AddBucket' procedure and handling errors.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.AddBucket.String(), req, res); err != nil {
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

func (h *handlers) RemoveBucketHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requestEventFactory.CreateRequestEvent(w, r)
	if err == client.ErrInvalidProtocol {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}
