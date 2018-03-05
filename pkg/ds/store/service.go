package store

import (
	"github.com/chanyoung/nil/pkg/ds/store/request"
	"github.com/chanyoung/nil/pkg/ds/store/volume"
)

// Service is a backend store interface.
type Service interface {
	// Run starts to service backend store.
	Run()

	// AddVolume adds a new volume into the store list.
	AddVolume(v *volume.Vol) error

	// Push pushes an io request into the scheduling queue.
	Push(c *request.Request) error
}
