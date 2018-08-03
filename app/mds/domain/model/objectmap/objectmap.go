package objectmap

import (
	"github.com/chanyoung/nil/pkg/omap"
)

// Repository provides to access object map database.
type Repository interface {
	FindByObject(object, size, bucket omap.Name) (*omap.OMap, error)
	FindByChunk(chunk omap.Name) (*omap.OMap, error)
	Update(...*omap.OMap) error
}
