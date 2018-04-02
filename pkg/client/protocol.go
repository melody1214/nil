package client

type Protocol string

func (p Protocol) String() string {
	return string(p)
}

const (
	S3      Protocol = "s3"
	Unknown          = "unknown"
)
