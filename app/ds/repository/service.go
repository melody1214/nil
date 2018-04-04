package repository

// Service is a backend store interface.
type Service interface {
	// Run starts to service backend store.
	Run()

	// Stop clean up the backend store and make it stop gracefully.
	Stop()

	// AddVolume adds a new volume into the store list.
	AddVolume(v *Vol) error

	// Push pushes an io request into the scheduling queue.
	Push(c *Request) error
}