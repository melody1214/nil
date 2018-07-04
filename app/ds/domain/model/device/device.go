package device

import "errors"

// ErrDeviceIsBusy is used when the device is busy.
var ErrDeviceIsBusy = errors.New("device is busy")

// ErrDeviceNotExists is used when the device not exists with the given path.
var ErrDeviceNotExists = errors.New("device not exists")

// ErrPermissionIsNotAllowed is used when the permission is not allowed.
var ErrPermissionIsNotAllowed = errors.New("permission is not allowed")

// ErrInvalidDevice is used when the device is not valid.
var ErrInvalidDevice = errors.New("invalid device")

// Device is the physical block device.
type Device struct {
	name  Name
	State State
}

// Name returns the device name.
func (d *Device) Name() Name {
	return d.name
}

// New creates a new block device object with the given name.
func New(name Name) *Device {
	return &Device{
		name: name,
	}
}

// Name represents the device name such as '/dev/sda1'.
type Name string

// IsValid returns true if the name is valid.
func (n Name) IsValid() bool {
	// HasPrefix /dev/sd..
	return true
}

// State represents the current state of disk.
type State string

const (
	// Active == spin up.
	Active State = "Active"
	// StandBy == spin down.
	StandBy State = "StandBy"
)

func (s State) String() string {
	return string(s)
}

// Repository provides to access device objects.
type Repository interface {
	Create(*Device) error
}
