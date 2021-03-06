package nilrpc

import (
	"github.com/chanyoung/nil/app/mds/application/gencoding/token"
	"github.com/chanyoung/nil/pkg/cmap"
)

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
	Token token.Token
}

type DGEEncodeResponse struct {
}

type DGEGetCandidateChunkRequest struct {
	Vol cmap.ID
	EG  cmap.ID
}

type DGEGetCandidateChunkResponse struct {
	Chunk string
}
