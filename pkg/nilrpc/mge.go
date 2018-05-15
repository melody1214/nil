package nilrpc

// MGEGGGRequest requests to generate global encoding group
// with the given regions
type MGEGGGRequest struct {
	Regions []string
}

// MGEGGGResponse returns the result of the request.
type MGEGGGResponse struct{}
