package recovery

import "github.com/chanyoung/nil/pkg/cmap"

type EncodingGroup struct {
	cmap.EncodingGroup
	firstVol  cmap.ID
	secondVol cmap.ID
	thirdVol  cmap.ID
	parityVol cmap.ID
}
