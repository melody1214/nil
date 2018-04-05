package consensus

import "github.com/chanyoung/nil/pkg/nilmux"

// Repository provides access to consensus protocol object.
type Repository interface {
	Open(raftL *nilmux.Layer) error
	Close() error
}
