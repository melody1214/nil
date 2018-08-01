package clustermap

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides to access cluster map database.
type Repository interface {
	UpdateWhole(*cmap.CMap) (*cmap.CMap, error)
	UpdateNode(*cmap.Node) (*cmap.CMap, error)
	FindLatest() (*cmap.CMap, error)
}
