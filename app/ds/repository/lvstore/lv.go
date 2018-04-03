package lvstore

import "github.com/chanyoung/nil/app/ds/repository"

type lv struct {
	// Embed volume.
	*repository.Vol

	// TODO: adds lv specific fields here.
}
