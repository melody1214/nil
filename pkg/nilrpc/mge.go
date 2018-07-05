package nilrpc

import "github.com/chanyoung/nil/app/mds/application/gencoding/token"

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

type MGEHandleTokenRequest struct {
	token.Token
}

type MGEHandleTokenResponse struct{}

type MGEGetEncodingJobRequest struct {
	Region string
}

type MGEGetEncodingJobResponse struct {
	Token token.Token
}

type MGESetJobStatusRequest struct {
	ID     int64
	Status int
}

type MGESetJobStatusResponse struct {
}

type MGEJobFinishedRequest struct {
	Token token.Token
}

type MGEJobFinishedResponse struct{}

type MGESetPrimaryChunkRequest struct {
	Primary token.Unencoded
	Job     int64
}

type MGESetPrimaryChunkResponse struct {
}
