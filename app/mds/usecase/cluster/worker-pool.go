package cluster

// WorkerPool is a service for managing worker. It is involved in two entities,
// a job entity and a worker entity. It fetches a job from the job repository
// and dispatches the worker.
type workerPool struct{}

// newWorkerPool returns a new worker pool service.
func newWorkerPool() *workerPool {
	return &workerPool{}
}
