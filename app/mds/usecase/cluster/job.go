package cluster

// id is an unique identifier for entity objects.
type id int64

func (i id) Int64() int64 { return int64(i) }

// jobState represents the current state of the job.
type jobState int

const (
	ready jobState = iota
	run
	done
)

// Job act as a kind of event history in the cluster domain.
// When a significant change occurs in configuring the cluster, the handler
// receiving the event generates a job through the job factory. Each job has
// an unique ID, in other words it is an entity. Job is distinguished by an
// its ID and can be changed its value by successive events.
type job struct {
	id    id
	event event
}
