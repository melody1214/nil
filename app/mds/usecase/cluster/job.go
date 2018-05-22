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

// String returns the string type of its value.
func (s JobState) String() string {
	if s == Ready {
		return "Ready"
	} else if s == Run {
		return "Run"
	} else if s == Done {
		return "Done"
	} else if s == Abort {
		return "Abort"
	} else if s == Merged {
		return "Merged"
	}
	return "Unknown"
}

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

// String returns the string type of its value.
func (t JobType) String() string {
	if t == Iterative {
		return "Iterative"
	} else if t == Batch {
		return "Batch"
	}
	return "Unknown"
}

// JobLog store the log about the job within 32 bytes long.
type JobLog string

func newJobLog(s string) JobLog {
	if len(s) < 64 {
		return JobLog(s)
	}
	return JobLog(string(s[:64]))
}

// String returns the string type of its value.
func (l JobLog) String() string {
	return string(l)
}

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
	Log         JobLog

	// This is a private fields for iterative type of jobs.
	private     interface{}
	waitChannel chan error

	// Store errors that occur during execution.
	err error

	// Set when the map is changed.
	mapChanged bool
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
