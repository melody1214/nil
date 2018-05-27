package nilrpc

import "github.com/chanyoung/nil/pkg/cmap"

type MOBObjectPutRequest struct {
	Name          string
	Bucket        string
	EncodingGroup cmap.ID
	Volume        cmap.ID
}
type MOBObjectPutResponse struct{}

type MOBObjectGetRequest struct {
	Name   string
	Bucket string
}
type MOBObjectGetResponse struct {
	EncodingGroupID cmap.ID
	VolumeID        cmap.ID
	DsID            cmap.ID
}

type MOBGetChunkRequest struct {
	EncodingGroup cmap.ID
}
type MOBGetChunkResponse struct {
	ID string
}

type MOBSetChunkRequest struct {
	Chunk         string
	EncodingGroup cmap.ID
	Status        string
}
type MOBSetChunkResponse struct {
}
