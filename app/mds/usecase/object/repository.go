package object

import "github.com/chanyoung/nil/pkg/cmap"

// Repository provides access to object database.
type Repository interface {
	Put(o *ObjInfo) error
	Get(name string) (*ObjInfo, error)
	GetChunk(eg cmap.ID) (cID string, err error)
}
