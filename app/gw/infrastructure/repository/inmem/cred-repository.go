package inmem

import (
	"sync"

	"github.com/chanyoung/nil/app/gw/domain/model/cred"
)

// credRepositoryInmem implements of cred repository.
type credRepositoryInmem struct {
	cache map[cred.Key]*cred.Cred
	mu    sync.Mutex
}

// Store put the credential into the inmem repository.
func (r *credRepositoryInmem) Store(c *cred.Cred) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c.IsExpired() {
		return
	}

	r.cache[c.AccessKey()] = c
}

// Find finds the matched credential with the given access key.
func (r *credRepositoryInmem) Find(access cred.Key) (*cred.Cred, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cache[access]
	if ok == false {
		return nil, cred.ErrNoSuchCred
	}

	if c.IsExpired() {
		delete(r.cache, access)
		return nil, cred.ErrNoSuchCred
	}

	return c, nil
}

// NewCredRepository returns a new instance of a in-memory cred repository.
func NewCredRepository() cred.Repository {
	return &credRepositoryInmem{
		cache: make(map[cred.Key]*cred.Cred),
	}
}
