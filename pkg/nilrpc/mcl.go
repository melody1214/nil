package nilrpc

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

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

// MCLLocalJoinRequest requests to join the local IDC cluster.
type MCLLocalJoinRequest struct {
	Node cmap.Node
}

// MCLLocalJoinResponse reponse the result of join request.
type MCLLocalJoinResponse struct {
}

// MCLGlobalJoinRequest includes an information for joining a new node into the raft clsuter.
// RaftAddr: address of the requested node.
// NodeID: ID of the requested node.
type MCLGlobalJoinRequest struct {
	RaftAddr string
	NodeID   string
}

// MCLGlobalJoinResponse is a NilRPC response message to join an existing cluster.
type MCLGlobalJoinResponse struct{}

// MCLRegisterVolumeRequest contains a new volume information.
type MCLRegisterVolumeRequest struct {
	Volume cmap.Volume
	// ID     string
	// Ds     string
	// Size   uint64
	// Free   uint64
	// Used   uint64
	// Speed  string
	// Status string
}

// MCLRegisterVolumeResponse contains a registered volume id and the results.
type MCLRegisterVolumeResponse struct {
	ID string
}

// MCLListJobRequest requests to show the job history.
type MCLListJobRequest struct{}

// MCLListJobResponse contains the job list.
type MCLListJobResponse struct {
	List []string
}
