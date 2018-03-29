package inmem

import (
	"sync"
	"time"

	"github.com/chanyoung/nil/app/gw/usecase/auth"
	"github.com/chanyoung/nil/pkg/s3"
)

const defaultCacheExpireTime = 10 * time.Minute

type authCache struct {
	key    s3.SecretKey
	expire time.Time
}

func (c *authCache) expired() bool {
	return time.Now().After(c.expire)
}

type authRepository struct {
	creds map[s3.AccessKey]authCache
	mu    sync.Mutex
}

func (r *authRepository) Find(accessKey s3.AccessKey) (s3.SecretKey, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.creds[accessKey]
	if ok == false {
		return s3.SecretKey(""), false
	}

	if c.expired() {
		delete(r.creds, accessKey)
		return s3.SecretKey(""), false
	}

	return c.key, true
}

func (r *authRepository) Add(accessKey s3.AccessKey, secretKey s3.SecretKey) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.creds[accessKey] = authCache{
		key:    secretKey,
		expire: time.Now().Add(defaultCacheExpireTime),
	}
}

// NewAuthRepository returns a new instance of a in-memory auth repository.
func NewAuthRepository() auth.Repository {
	return &authRepository{
		creds: make(map[s3.AccessKey]authCache),
	}
}
