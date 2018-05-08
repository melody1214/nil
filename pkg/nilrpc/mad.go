package nilrpc

// MADJoinRequest includes an information for joining a new node into the raft clsuter.
// RaftAddr: address of the requested node.
// NodeID: ID of the requested node.
type MADJoinRequest struct {
	RaftAddr string
	NodeID   string
}

// MADJoinResponse is a NilRPC response message to join an existing cluster.
type MADJoinResponse struct{}

// MADAddUserRequest requests to create a new user with the given name.
type MADAddUserRequest struct {
	Name string
}

// MADAddUserResponse response AddUserRequest with the AccessKey and SecretKey.
type MADAddUserResponse struct {
	AccessKey string
	SecretKey string
}

// MADRegisterVolumeRequest contains a new volume information.
type MADRegisterVolumeRequest struct {
	ID     string
	Ds     string
	Size   uint64
	Free   uint64
	Used   uint64
	Speed  string
	Status string
}

// MADRegisterVolumeResponse contains a registered volume id and the results.
type MADRegisterVolumeResponse struct {
	ID string
}
