package nilrpc

import "github.com/chanyoung/nil/pkg/s3"

// MBUMakeBucketRequest requests to create bucket for given name and user.
type MBUMakeBucketRequest struct {
	BucketName string
	AccessKey  string
	Region     string
}

// MBUMakeBucketResponse responses the result of addBucket.
type MBUMakeBucketResponse struct {
	S3ErrCode s3.ErrorCode
}
