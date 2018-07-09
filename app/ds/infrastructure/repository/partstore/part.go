package partstore

import "github.com/chanyoung/nil/app/ds/domain/model/volume"

type vol struct {
	// Embed volume.
	*volume.Volume

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
