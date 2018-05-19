package nilrpc

import "github.com/chanyoung/nil/pkg/cmap"

type DGERenameChunkRequest struct {
	Vol      string
	EncGrp   string
	OldChunk string
	NewChunk string
}

type DGERenameChunkResponse struct {
}

type DGETruncateChunkRequest struct {
	Vol    string
	EncGrp string
	Chunk  string
}

type DGETruncateChunkResponse struct {
}

type DGEEncodeRequest struct {
	FirstRegion  string
	SecondRegion string
	ThirdRegion  string
	TblID        int
}

type DGEEncodeResponse struct {
}

type DGEPrepareEncodingRequest struct {
	Vol cmap.ID
	EG  cmap.ID
}

type DGEPrepareEncodingResponse struct {
	Chunk string
}
