package mds

// Config holds info required to set a metadata server.
type Config struct {
	// ServerAddr is the address of the metadata server.
	ServerAddr string

	// ServerPort is the port of the metadata server.
	ServerPort string

	// LogLocation is the file path of mds logging.
	// If it is empty, logging message will print out to stdout.
	LogLocation string
}
