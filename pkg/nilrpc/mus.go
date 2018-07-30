package nilrpc

import "github.com/chanyoung/nil/pkg/s3"

// MUSAddUserRequest requests to create a new user with the given name.
type MUSAddUserRequest struct {
	Name string
}

// MUSAddUserResponse response AddUserRequest with the AccessKey and SecretKey.
type MUSAddUserResponse struct {
	AccessKey string
	SecretKey string
}

// MUSGetCredentialRequest requests a credential for the given access key.
type MUSGetCredentialRequest struct {
	AccessKey string
}

// MUSGetCredentialResponse response the credential.
type MUSGetCredentialResponse struct {
	Exist     bool
	AccessKey string
	SecretKey string
}

// MUSMakeBucketRequest requests to create bucket for given name and user.
type MUSMakeBucketRequest struct {
	BucketName string
	AccessKey  string
	Region     string
}

// MUSMakeBucketResponse responses the result of addBucket.
type MUSMakeBucketResponse struct {
	S3ErrCode s3.ErrorCode
}
