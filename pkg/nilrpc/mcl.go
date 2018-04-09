package nilrpc

// MCLGetClusterMapRequest requests to get local cluster map.
// Version == 0; requests the latest version.
// Version > 0; requests higher version than given version.
type MCLGetClusterMapRequest struct {
	Version int64
}

// ClusterNode represents the nodes.
type ClusterNode struct {
	ID   int64
	Name string
	Addr string
	Type string
	Stat string
}

// MCLGetClusterMapResponse contains a current local cluster members.
type MCLGetClusterMapResponse struct {
	Version int64
	Nodes   []ClusterNode
}

// MCLGetUpdateNotiRequest requests to receive notification
// when the cluster map is updated. Gives some notification if
// has higher than given version of cluster map.
type MCLGetUpdateNotiRequest struct {
	Version int64
}

// MCLGetUpdateNotiResponse will response the cluster map is updated.
type MCLGetUpdateNotiResponse struct{}
