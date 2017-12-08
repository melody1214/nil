package config

// Raft holds info required to set a raft server.
type Raft struct {
	// LocalClusterRigion is the region name of the local cluster.
	LocalClusterRegion string
	// LocalClusterAddr is the endpoint address of the local cluster.
	LocalClusterAddr string

	// GlobalClusterAddr is one of the endpoint address of raft cluster.
	// Mds will ask here to try to join the raft clsuter.
	GlobalClusterAddr string
	// ClusterJoin is set 'true' when mds is going to join existed raft
	// cluster. Set 'false' when this is the first mds of whole cluster.
	ClusterJoin string

	// ElectionTimeout : Follower didn't receives a heartbeat message
	// over a 'election timeout' period, then it starts new election term.
	ElectionTimeout string
}
