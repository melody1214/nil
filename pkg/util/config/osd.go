package config

// Osd holds info required to set a object storage daemon.
type Osd struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string
	// ServerPort is the port of the metadata server.
	ServerPort string

	// Swim config.
	Swim Swim
	// Security config.
	Security Security

	// LogLocation is the file path of osd logging.
	// Default output path is stderr.
	LogLocation string
}
