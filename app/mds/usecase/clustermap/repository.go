package clustermap

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides access to cluster map database.
type Repository interface {
	GetClusterMapNodes() ([]cmap.Node, error)
	GetNewClusterMapVer() (cmap.Version, error)
}
