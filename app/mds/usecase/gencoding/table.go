package gencoding

// Status means the global encoding job status
type Status int

const (
	// Ready means ready to encode.
	Ready Status = iota
	// Run means now on encoding.
	Run
	// Finish means encoding is done.
	Finish
)
