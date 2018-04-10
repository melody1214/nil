package cmap

// VolumeSpeed represents a disk speed level.
type VolumeSpeed int

const (
	// Low : 60 Mb/s < speed < 80 Mb/s
	Low VolumeSpeed = iota
	// Mid : 80 Mb/s < speed < 100 Mb/s
	Mid
	// High : 100 Mb/s < speed
	High
)

func (s VolumeSpeed) String() string {
	switch s {
	case Low:
		return "low"
	case Mid:
		return "mid"
	case High:
		return "high"
	default:
		return "unknown"
	}
}

// VolumeStatus represents a disk status.
type VolumeStatus int

const (
	// Prepared represents the volume is ready to run.
	Prepared VolumeStatus = iota
	// Active represents the volume is now running.
	Active
	// Failed represents the volume has some problems and stopped now.
	Failed
)

func (s VolumeStatus) String() string {
	switch s {
	case Prepared:
		return "prepared"
	case Active:
		return "active"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

// Volume is volumes which is attached in the ds.
type Volume struct {
	ID      ID
	Size    uint64
	Free    uint64
	Used    uint64
	Speed   VolumeSpeed
	Status  VolumeStatus
	Nodes   []ID
	EncGrps []ID
}
