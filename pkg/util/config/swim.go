package config

// Swim includes info required to set a swim server.
type Swim struct {
	// CoordinatorAddr is the address of the swim node
	// which will ask to join the cluster.
	CoordinatorAddr string

	// Period is an interval time of pinging.
	Period string

	// Expire is an expire time of pinging.
	Expire string
}
