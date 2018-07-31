package config

// Mds holds info required to set a metadata server.
type Mds struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string
	// ServerPort is the port of the metadata server.
	ServerPort string

	// WorkDir is a working directory of the mds.
	WorkDir string

	// Rebalance is a period of check balance.
	Rebalance string

	// MySQLUser is the user ID of MySQL database.
	MySQLUser string
	// MySQLPassword is the password of MySQL user.
	MySQLPassword string
	// MySQLDatabase is the schema name.
	MySQLDatabase string
	// MySQLHost is the host address of MySQL server.
	MySQLHost string
	// MySQLPort is the port number of MySQL server.
	MySQLPort string

	// Raft config.
	Raft Raft
	// Swim config.
	Swim Swim
	// Security config.
	Security Security

	// Encoding
	// LocalEncodingMatrices is the number of matrices for local encoding.
	LocalEncodingMatrices string
	// GlobalEncodingMatrices is the number of matrices for global encoding.
	GlobalEncodingMatrices string

	// LogLocation is the file path of mds logging.
	// Default output path is stderr.
	LogLocation string
}
