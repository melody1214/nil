package cluster

// jobFactory is a factory for making job.
// It takes an event as an argument and creates a job with a unique ID.
// Created jobs are stored in the job repository.
type jobFactory struct {
	store jobRepository
}

// newJobFactory returns a new job factory object.
func newJobFactory(s jobRepository) *jobFactory {
	return &jobFactory{
		store: s,
	}
}

// create creates an event with a given event information.
func (f *jobFactory) create(e *Event) error {
	return nil
}
