package user

import (
	"github.com/chanyoung/nil/pkg/security"
)

// Repository provides access to admin database.
type Repository interface {
	AmILeader() (bool, error)
	LeaderEndpoint() (endpoint string)
	AddUser(name string, ak security.APIKey) error
	MakeBucket(bucketName, accessKey, region string) (err error)
	FindSecretKey(accessKey string) (secretKey string, err error)
}
