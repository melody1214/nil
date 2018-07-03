package nilrpc

import "github.com/chanyoung/nil/pkg/cmap"

type DOBSetChunkPoolRequest struct {
	ID    string
	EG    cmap.ID
	Vol   cmap.ID
	Shard int
}

type DOBSetChunkPoolResponse struct {
}
