package repository

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

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

// Status represents a disk status.
type Status int

const (
	// Prepared represents the volume is ready to run.
	Prepared Status = iota
	// Active represents the volume is now running.
	Active
	// Failed represents the volume has some problems and stopped now.
	Failed
)

func (s Status) String() string {
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

// ObjMap contains mapping information of objects.
type ObjMap struct {
	Cid    string
	Offset int64
}

// Vol contains information about the volume.
type Vol struct {
	Name     string
	Dev      string
	MntPoint string
	Size     uint64
	Free     uint64
	Used     uint64
	Speed    VolumeSpeed
	Status   Status

	ChunkSize int64
	Objs      map[string]ObjMap
}

// NewVol collects information about the volume with the given
// device path and returns a pointer of Vol.
func NewVol(dev string) (v *Vol, err error) {
	// Get absolute device path.
	if dev, err = filepath.Abs(dev); err != nil {
		return nil, err
	}

	// Creates the LV with the given device path.
	v = &Vol{
		Dev:    dev,
		Status: Prepared,
		Objs:   make(map[string]ObjMap),
	}

	// Checks the given device path is valid.
	if err = v.CheckDevicePath(); err != nil {
		return nil, err
	}

	return
}

// CheckDevicePath checks the device of logical volume exists.
func (v *Vol) CheckDevicePath() error {
	if _, err := os.Lstat(v.Dev); os.IsNotExist(err) {
		return fmt.Errorf("device with the given path is not exist")
	} else if err != nil {
		return fmt.Errorf("device with the given path is not valid: %v", err)
	}

	return nil
}

// Mount mounts the device to the target directory.
func (v *Vol) Mount() (err error) {
	os.Mkdir(v.MntPoint, 0775)

	if v.MntPoint, err = filepath.Abs(v.MntPoint); err != nil {
		return err
	}

	if output, err := exec.Command("mount", v.Dev, v.MntPoint).CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %v", output, err)
	}

	return nil
}

// Umount unmounts the device from the mount point directory.
func (v *Vol) Umount() error {
	output, err := exec.Command("umount", v.MntPoint).CombinedOutput()
	if err == nil {
		return nil
	}

	output, err = exec.Command("umount", v.Dev).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s: %v", output, err)
	}

	return err
}

// UpdateStatFs updates the stat of the device.
func (v *Vol) UpdateStatFs() error {
	fs := syscall.Statfs_t{}
	// if err := syscall.Statfs(l.Dev, &fs); err != nil {
	if err := syscall.Statfs(v.MntPoint, &fs); err != nil {
		return err
	}

	v.Size = fs.Blocks * uint64(fs.Bsize) / 1024 / 1024
	v.Free = fs.Bfree * uint64(fs.Bsize) / 1024 / 1024
	v.Used = v.Size - v.Free
	return nil
}

// Usage returns a disk usage of this volume.
func (v *Vol) Usage() int {
	v.UpdateStatFs()

	if v.Size < 1 {
		return 0
	}

	return int((v.Used * 100) / v.Size)
}

// SetSpeed set a disk speed level of this volume.
func (v *Vol) SetSpeed() {
	speeds := [3]VolumeSpeed{High, Mid, Low}

	v.Speed = speeds[rand.Intn(len(speeds))]
}
