package clustermap

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides to access cluster map database.
type Repository interface {
	CreateInitial() *cmap.CMap
	Update(*cmap.CMap) (*cmap.CMap, error)
	FindLatest() *cmap.CMap
}
