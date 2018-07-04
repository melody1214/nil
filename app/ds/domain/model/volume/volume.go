package volume

import (
	"errors"
	"sync"
)

// ErrNotFound is used when the find operation failed.
var ErrNotFound = errors.New("failed to find the volume with the given name")

// Volume contains information about the volum.
type Volume struct {
	name     Name
	mntPoint string
	speed    Speed
	status   Status
	size     uint64
	lock     sync.RWMutex
}

// SetStatus set the status with the given value.
func (v *Volume) SetStatus(new Status) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.status = new
}

// SetSize set the volume size with the given amount.
func (v *Volume) SetSize(size uint64) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.size = size
}

// Name returns a volume name.
func (v *Volume) Name() Name {
	return v.name
}

// MntPoint returns a volume mount point.
func (v *Volume) MntPoint() string {
	return v.mntPoint
}

// Speed returns a volume speed.
func (v *Volume) Speed() Speed {
	return v.speed
}

// Status returns a volume status.
func (v *Volume) Status() Status {
	v.lock.RLock()
	defer v.lock.RUnlock()

	return v.status
}

// Size returns a volume size.
func (v *Volume) Size() uint64 {
	v.lock.RLock()
	defer v.lock.RUnlock()

	return v.size
}

// New creates the new volume object.
func New(name Name, mntPoint string, speed Speed, size uint64) *Volume {
	return &Volume{
		name:     name,
		mntPoint: mntPoint,
		speed:    speed,
		status:   Active,
		size:     size,
	}
}

// Name represents the device name such as '/dev/sda1'.
type Name string

func (n Name) String() string {
	return string(n)
}

// Speed represents a disk speed level.
type Speed string

const (
	// Low : 60 Mb/s < speed < 80 Mb/s
	Low Speed = "Low"
	// Mid : 80 Mb/s < speed < 100 Mb/s
	Mid = "Mid"
	// High : 100 Mb/s < speed
	High = "High"
)

func (s Speed) String() string {
	switch s {
	case Low, Mid, High:
		return string(s)
	default:
		return "unknown"
	}
}

// Status represents a disk status.
type Status string

const (
	// Active represents the volume is now running.
	Active = "Active"
	// Failed represents the volume has some problems and stopped now.
	Failed = "Failed"
)

func (s Status) String() string {
	switch s {
	case Active, Failed:
		return string(s)
	default:
		return "unknown"
	}
}

// Repository provides to access volume objects.
type Repository interface {
	Find(name Name) (Volume, error)
	FindAll() []Volume
}
