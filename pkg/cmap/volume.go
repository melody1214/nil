package cmap

// VolumeSpeed represents a disk speed level.
type VolumeSpeed string

const (
	// Low : 60 Mb/s < speed < 80 Mb/s
	Low VolumeSpeed = "Low"
	// Mid : 80 Mb/s < speed < 100 Mb/s
	Mid VolumeSpeed = "Mid"
	// High : 100 Mb/s < speed
	High VolumeSpeed = "High"
)

func (s VolumeSpeed) String() string {
	switch s {
	case Low, Mid, High:
		return string(s)
	default:
		return unknown
	}
}

// VolumeStatus represents a disk status.
type VolumeStatus string

const (
	// Prepared represents the volume is ready to run.
	Prepared VolumeStatus = "Prepared"
	// Active represents the volume is now running.
	Active VolumeStatus = "Active"
	// Failed represents the volume has some problems and stopped now.
	Failed VolumeStatus = "Failed"
)

func (s VolumeStatus) String() string {
	switch s {
	case Prepared, Active, Failed:
		return string(s)
	default:
		return unknown
	}
}

// Volume is volumes which is attached in the ds.
type Volume struct {
	ID      ID           `xml:"id"`
	Incr    Incarnation  `xml:"incarnation"`
	Size    uint64       `xml:"size"`
	Speed   VolumeSpeed  `xml:"speed"`
	Stat    VolumeStatus `xml:"status"`
	Node    ID           `xml:"node"`
	EncGrps []ID         `xml:"encgrp"`
	MaxEG   int          `xml:"maxeg"`
}
