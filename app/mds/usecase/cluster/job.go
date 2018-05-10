package cluster

import (
	"fmt"
)

// ID is an unique identifier for entity objects.
type ID int64

// Int64 returns its value in built-in int64 type.
func (i ID) Int64() int64 { return int64(i) }

// JobState represents the current state of the job.
type JobState int

const (
	// Ready : ready to run.
	Ready JobState = iota
	// Run : now on running.
	Run
	// Done : finished.
	Done
	// Abort : Stop trying, it is impossible to perform.
	Abort
	// Merged : merged because the same event is occured before done.
	Merged
)

// JobType represents the type of job.
type JobType int

const (
	// Iterative : job requester is waiting for processing.
	// Scheduler try to dispatch the iterative job as soon as possible.
	Iterative JobType = iota
	// Batch : job requester is not waiting.
	// This kind of job take a quite long time, for instance fail over.
	Batch
)

// Job act as a kind of event history in the cluster domain.
// When a significant change occurs in configuring the cluster, the handler
// receiving the event generates a job through the job factory. Each job has
// an unique ID, in other words it is an entity. Job is distinguished by an
// its ID and can be changed its value by successive events.
type Job struct {
	ID          ID
	Type        JobType
	State       JobState
	Event       Event
	ScheduledAt Time
	FinishedAt  Time

	// This is a private fields for iterative type of jobs.
	private     interface{}
	waitChannel chan error
}

func (j *Job) getPrivate() (interface{}, error) {
	if j.Type != Iterative {
		return nil, fmt.Errorf("only iterative type of job has private field")
	}
	return j.private, nil
}

func (j *Job) getWaitChannel() (chan error, error) {
	if j.Type != Iterative {
		return nil, fmt.Errorf("only iterative type of job has wait channel field")
	}
	return j.waitChannel, nil
}
