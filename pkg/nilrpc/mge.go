package nilrpc

// MGEGGGRequest requests to generate global encoding group
// with the given regions
type MGEGGGRequest struct {
	Regions []string
}

// MGEGGGResponse returns the result of the request.
type MGEGGGResponse struct{}

type MGEUpdateUnencodedChunkRequest struct {
	Region    string
	Unencoded int
}

type MGEUpdateUnencodedChunkResponse struct{}

type MGESelectEncodingGroupRequest struct {
	TblID int64
}

type MGESelectEncodingGroupResponse struct{}
