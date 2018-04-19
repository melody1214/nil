package recovery

type jobID int

// intent represents the intent of the jobs.
type intent int

// recoveryJob is allocated per one failure encoding group, or object.
// Worker can make job and dispatch it, and monitoring jobs for fine-
// grained recovery control.
type recoveryJob struct {
	id jobID
}

const (
	// new means the job will make new encoding groups.
	newEncodingGroup intent = iota
	// move means the job will move existing encoding group
	// members to other ds for rebalance the cluster.
	moveEncodingGroup
)

// rebalanceJob is allocated per one encoding group.
// Controlled by worker.
type rebalanceJob struct {
	id     jobID
	intent intent
}
