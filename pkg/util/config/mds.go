package config

// Mds holds info required to set a metadata server.
type Mds struct {
	// ID is the uuid of the metadata server.
	ID string

	// ServerAddr is the address of the metadata server.
	ServerAddr string
	// ServerPort is the port of the metadata server.
	ServerPort string

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

	// LocalClusterRigion is the region name of the local cluster.
	LocalClusterRegion string
	// LocalClusterAddr is the endpoint address of the local cluster.
	LocalClusterAddr string

	// Raft config.
	Raft Raft
	// Security config.
	Security Security

	// LogLocation is the file path of mds logging.
	// Default output path is stderr.
	LogLocation string
}
