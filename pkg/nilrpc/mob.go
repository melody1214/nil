package nilrpc

type MOBObjectPutRequest struct {
	Name          string
	Bucket        string
	EncodingGroup string
	Volume        string
}
type MOBObjectPutResponse struct{}

type MOBObjectGetRequest struct {
	Name   string
	Bucket string
}
type MOBObjectGetResponse struct {
	EncodingGroupID int64
	VolumeID        int64
	DsID            int64
}
