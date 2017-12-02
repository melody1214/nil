package config

// Gw holds info required to set a gateway server.
type Gw struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string
	// ServerPort is the port of the metadata server.
	ServerPort string
	// FirstMds is the mds address which will be used in the first contact
	// for getting local cluster membership information.
	FirstMds string

	// LogLocation is the file path of mds logging.
	// Default output path is stderr.
	LogLocation string

	// UseHTTPS uses https to communicate client applications.
	UseHTTPS string
	// Security is the container of the information related with security.
	Security Security
}
