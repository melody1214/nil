package client

// Protocol is the type of the client request protocols.
type Protocol string

func (p Protocol) String() string {
	return string(p)
}

const (
	// S3 : Amazon S3
	S3 Protocol = "s3"
	// Unknown : unknown, not supported yet or invalid.
	Unknown = "unknown"
)
