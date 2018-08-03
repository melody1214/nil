package omap

import (
	"github.com/chanyoung/nil/pkg/cmap"
)

// OMap contains the information for looking object or chunk.
type OMap struct {
	Object Name
	Chunk  Name
	DS     cmap.ID
	DSs    []cmap.ID

	CurrentLv Level
	FinalLv   Level

	CMapVer cmap.Version
}

// Name is the name of object.
type Name string

func (n Name) String() string {
	return string(n)
}

type Level int

const (
	Replicated Level = iota
	LocalEncoded
	LocalEncodedMerged
	GlobalEncoded
)
