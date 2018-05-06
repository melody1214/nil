package recovery

import (
	"github.com/chanyoung/nil/pkg/cluster"
)

type EncodingGroup struct {
	cluster.EncodingGroup
	firstVol  cluster.ID
	secondVol cluster.ID
	thirdVol  cluster.ID
	parityVol cluster.ID
}
