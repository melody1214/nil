package clustermap

import (
	"database/sql"

	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides access to cluster map database.
type Repository interface {
	FindAllNodes() ([]cmap.Node, error)
	GetNewClusterMapVer() (cmap.Version, error)
	Begin() (*sql.Tx, error)
}
