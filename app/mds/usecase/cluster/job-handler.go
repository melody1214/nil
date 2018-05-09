package cluster

// jobFactory is a factory for making job.
// It takes an event as an argument and creates a job with a unique ID.
// Created jobs are stored in the job repository.
type jobFactory struct{}

// newJobFactory returns a new job factory object.
func newJobFactory() *jobFactory {
	return &jobFactory{}
}

// create creates an event with a given event information.
func (f *jobFactory) create(e event) error {
	return nil
}
