package config

// Ds holds info required to set a object storage daemon.
type Ds struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string
	// ServerPort is the port of the metadata server.
	ServerPort string

	// WorkDir is a working directory of the ds.
	WorkDir string

	// Store is the type of backend store.
	Store string
	// ChunkSize is the size of backend store chunk.
	ChunkSize string

	// Encoding
	// LocalParityShards is the number of shards that used to generate local parity.
	LocalParityShards string

	// Swim config.
	Swim Swim
	// Security config.
	Security Security

	// LogLocation is the file path of ds logging.
	// Default output path is stderr.
	LogLocation string
}
