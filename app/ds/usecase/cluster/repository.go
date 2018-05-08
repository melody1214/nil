package cluster

import "github.com/chanyoung/nil/app/ds/repository"

// Repository provides an persistent object store.
type Repository interface {
	// AddVolume adds a new volume into the store list.
	AddVolume(v *repository.Vol) error
}
