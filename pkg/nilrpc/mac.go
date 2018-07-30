package nilrpc

import "github.com/chanyoung/nil/pkg/s3"

// MACAddUserRequest requests to create a new user with the given name.
type MACAddUserRequest struct {
	Name string
}

// MACAddUserResponse response AddUserRequest with the AccessKey and SecretKey.
type MACAddUserResponse struct {
	AccessKey string
	SecretKey string
}

// MACGetCredentialRequest requests a credential for the given access key.
type MACGetCredentialRequest struct {
	AccessKey string
}

// MACGetCredentialResponse response the credential.
type MACGetCredentialResponse struct {
	Exist     bool
	AccessKey string
	SecretKey string
}

// MACMakeBucketRequest requests to create bucket for given name and user.
type MACMakeBucketRequest struct {
	BucketName string
	AccessKey  string
	Region     string
}

// MACMakeBucketResponse responses the result of addBucket.
type MACMakeBucketResponse struct {
	S3ErrCode s3.ErrorCode
}
