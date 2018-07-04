package partstore

import (
	"github.com/chanyoung/nil/app/ds/infrastructure/repository"
)

type pg struct {
	// Embed volume.
	*repository.Vol

	// TODO: adds lv specific fields here.
}

// dev contains device information.
type dev struct {
	Name        string
	State       string
	ActiveTime  uint
	StandbyTime uint
	Timestamp   uint
	TotalIO     uint
	Size        uint
	Free        uint
	Used        uint
}
