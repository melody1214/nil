package nilrpc

import "github.com/chanyoung/nil/pkg/cmap"

// MCLGetClusterMapRequest requests to get local cluster map.
// Version == 0; requests the latest version.
// Version > 0; requests higher version than given version.
type MCLGetClusterMapRequest struct {
	Version int64
}

// MCLGetClusterMapResponse contains a current local cluster members.
type MCLGetClusterMapResponse struct {
	ClusterMap cmap.CMap
}

// MCLGetUpdateNotiRequest requests to receive notification
// when the cluster map is updated. Gives some notification if
// has higher than given version of cluster map.
type MCLGetUpdateNotiRequest struct {
	Version int64
}

// MCLGetUpdateNotiResponse will response the cluster map is updated.
type MCLGetUpdateNotiResponse struct{}

// MCLUpdateClusterMapRequest requests to update cluster map.
type MCLUpdateClusterMapRequest struct{}

// MCLUpdateClusterMapResponse includes the result of update cluster map.
type MCLUpdateClusterMapResponse struct{}
