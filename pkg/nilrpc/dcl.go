package nilrpc

// DCLAddVolumeRequest requests to add new volume with the given device path.
type DCLAddVolumeRequest struct {
	DevicePath string
}

// DCLAddVolumeResponse is a response message to add volume request.
type DCLAddVolumeResponse struct{}

// type DCLRecoveryChunkRequest struct {
// 	ChunkID     string
// 	ChunkStatus string
// 	ChunkEG     cmap.ID
// 	ChunkVol    cmap.ID
// 	TargetVol   cmap.ID
// 	Type        string
// }

// type DCLRecoveryChunkResponse struct{}
