package lvstore

import "github.com/chanyoung/nil/app/ds/store/volume"

type lv struct {
	// Embed volume.
	*volume.Vol

	// TODO: adds lv specific fields here.
}
