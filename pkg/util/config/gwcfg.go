package config

// Gw holds info required to set a gateway server.
type Gw struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string

	// ServerPort is the port of the metadata server.
	ServerPort string

	// LogLocation is the file path of mds logging.
	// Default output path is stderr.
	LogLocation string

	// UseHTTPS uses https to communicate client applications.
	UseHTTPS string

	// CertsDir is the directory path which contains files for https configuration.
	CertsDir string
}
