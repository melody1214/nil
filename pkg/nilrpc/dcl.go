package nilrpc

// DCLAddVolumeRequest requests to add new volume with the given device path.
type DCLAddVolumeRequest struct {
	DevicePath string
}

// DCLAddVolumeResponse is a response message to add volume request.
type DCLAddVolumeResponse struct{}
