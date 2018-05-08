package nilrpc

// DADAddVolumeRequest requests to add new volume with the given device path.
type DADAddVolumeRequest struct {
	DevicePath string
}

// DADAddVolumeResponse is a response message to add volume request.
type DADAddVolumeResponse struct{}
