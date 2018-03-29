package auth

import (
	"github.com/chanyoung/nil/pkg/s3"
)

// Repository provides an authentication cache.
type Repository interface {
	Find(s3.AccessKey) (s3.SecretKey, bool)
	Add(s3.AccessKey, s3.SecretKey)
}
