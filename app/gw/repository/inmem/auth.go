package inmem

import (
	"sync"
	"time"

	"github.com/chanyoung/nil/app/gw/usecase/auth"
)

const defaultCacheExpireTime = 10 * time.Minute

type authCache struct {
	secretKey string
	expire    time.Time
}

func (c *authCache) expired() bool {
	return time.Now().After(c.expire)
}

type authRepository struct {
	creds map[string]authCache
	mu    sync.Mutex
}

func (r *authRepository) Find(accessKey string) (secretKey string, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.creds[accessKey]
	if ok == false {
		return "", false
	}

	if c.expired() {
		delete(r.creds, accessKey)
		return "", false
	}

	return c.secretKey, true
}

func (r *authRepository) Add(accessKey string, secretKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.creds[accessKey] = authCache{
		secretKey: secretKey,
		expire:    time.Now().Add(defaultCacheExpireTime),
	}
}

// NewAuthRepository returns a new instance of a in-memory auth repository.
func NewAuthRepository() auth.Repository {
	return &authRepository{
		creds: make(map[string]authCache),
	}
}
