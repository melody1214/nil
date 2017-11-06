package mds

// Config holds info required to set a metadata server.
type Config struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string

	// ServerPort is the port of the metadata server.
	ServerPort string

	// LogLocation is the file path of mds logging.
	// Default output path is stderr.
	LogLocation string
}
