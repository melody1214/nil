package gencoding

// Status represents the job status of global encoding.
type Status int

const (
	// Ready means the job is ready.
	Ready Status = iota
	// Run means the job is currently working.
	Run
	// Fail means the job is failed.
	Fail
	// Done means the job is finished successfully.
	Done
)

func (s *service) encode() {

}
