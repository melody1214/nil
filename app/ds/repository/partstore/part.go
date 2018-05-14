package partstore

import "github.com/chanyoung/nil/app/ds/repository"

type pg struct {
	// Embed volume.
	*repository.Vol

	// TODO: adds lv specific fields here.
}
