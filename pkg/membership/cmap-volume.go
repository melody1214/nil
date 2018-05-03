package membership

// VolumeSpeed represents a disk speed level.
type VolumeSpeed string

const (
	// Low : 60 Mb/s < speed < 80 Mb/s
	Low VolumeSpeed = "Low"
	// Mid : 80 Mb/s < speed < 100 Mb/s
	Mid = "Mid"
	// High : 100 Mb/s < speed
	High = "High"
)

func (s VolumeSpeed) String() string {
	switch s {
	case Low, Mid, High:
		return string(s)
	default:
		return "Unknown"
	}
}

// VolumeStatus represents a disk status.
type VolumeStatus string

const (
	// Prepared represents the volume is ready to run.
	Prepared VolumeStatus = "Prepared"
	// Active represents the volume is now running.
	Active = "Active"
	// Failed represents the volume has some problems and stopped now.
	Failed = "Failed"
)

func (s VolumeStatus) String() string {
	switch s {
	case Prepared, Active, Failed:
		return string(s)
	default:
		return "Unknown"
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
	Node    ID
	EncGrps []ID
}
