package recovery

import "github.com/chanyoung/nil/pkg/cmap"

type localChain struct {
	cmap.EncodingGroup
	firstVol  cmap.ID
	secondVol cmap.ID
	thirdVol  cmap.ID
	parityVol cmap.ID
}
