package config

// Swim includes info required to set a swim server.
type Swim struct {
	// ClusterJoinAddr is the address of the swim node
	// which will ask to join the cluster.
	ClusterJoinAddr string

	// ID is the uuid of this swim node.
	ID string
	// Host is the host address of this swim node.
	Host string
	// Port is the port number of this swim node.
	Port string
	// Type represents this swim node is MDS or OSD.
	Type string

	// Security config.
	Security Security
}
