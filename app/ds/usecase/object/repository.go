package object

import "github.com/chanyoung/nil/app/ds/repository"

// Repository provides an persistent object store.
type Repository interface {
	// Push pushes an io request into the scheduling queue.
	Push(r *repository.Request) error
}
