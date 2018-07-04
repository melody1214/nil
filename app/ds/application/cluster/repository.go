package cluster

import "github.com/chanyoung/nil/app/ds/infrastructure/repository"

// Repository provides an persistent object store.
type Repository interface {
	// AddVolume adds a new volume into the store list.
	AddVolume(v *repository.Vol) error
	Push(r *repository.Request) error
	BuildObjectMap(Vol string, cid string) error
}
