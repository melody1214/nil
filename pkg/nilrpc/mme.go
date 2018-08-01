package nilrpc

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

// MMEGetClusterMapRequest requests to get local cluster map.
// Version == 0; requests the latest version.
// Version > 0; requests higher version than given version.
type MMEGetClusterMapRequest struct {
	Version int64
}

// MMEGetClusterMapResponse contains a current local cluster members.
type MMEGetClusterMapResponse struct {
	ClusterMap cmap.CMap
}

// MMEGetUpdateNotiRequest requests to receive notification
// when the cluster map is updated. Gives some notification if
// has higher than given version of cluster map.
type MMEGetUpdateNotiRequest struct {
	Version int64
}

// MMEGetUpdateNotiResponse will response the cluster map is updated.
type MMEGetUpdateNotiResponse struct{}

// MMELocalJoinRequest requests to join the local IDC cluster.
type MMELocalJoinRequest struct {
	Node cmap.Node
}

// MMELocalJoinResponse reponse the result of join request.
type MMELocalJoinResponse struct {
}

// MMEGlobalJoinRequest includes an information for joining a new node into the raft clsuter.
// RaftAddr: address of the requested node.
// NodeID: ID of the requested node.
type MMEGlobalJoinRequest struct {
	RaftAddr string
	NodeID   string
}

// MMEGlobalJoinResponse is a NilRPC response message to join an existing cluster.
type MMEGlobalJoinResponse struct{}
