package gencoding

import "github.com/chanyoung/nil/app/ds/repository"

// Repository provides an persistent object store.
type Repository interface {
	Push(r *repository.Request) error
	CountEncChunk(Vol string, LocGid string) (int, error)
	RenameChunk(src string, dest string, Vol string, LocGid string) error
}
