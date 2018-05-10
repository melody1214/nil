package cluster

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
)

// JobType represents the type of job.
type JobType int

const (
	// Interactive : job requester is waiting for processing.
	// Scheduler try to dispatch the interative job as soon as possible.
	Interactive JobType = iota
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

	// This is a private field for batch type of jobs.
	batchPrivate interface{}
}
