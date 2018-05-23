package repository

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

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
		return "unknown"
	}
}

// Status represents a disk status.
type Status string

const (
	// Prepared represents the volume is ready to run.
	Prepared Status = "Prepared"
	// Active represents the volume is now running.
	Active = "Active"
	// Failed represents the volume has some problems and stopped now.
	Failed = "Failed"
)

func (s Status) String() string {
	switch s {
	case Prepared, Active, Failed:
		return string(s)
	default:
		return "unknown"
	}
}

// ObjMap contains mapping information of objects and chunks.
type ObjMap struct {
	Cid     string
	Offset  int64
	ObjInfo ObjInfo
}

// ObjInfo contains information of objects.
type ObjInfo struct {
	Size int64
	MD5  string
}

// ChunkMap contains mapping information between chunks to partitions.
type ChunkMap struct {
	PartID    string
	ChunkInfo ChunkInfo
}

// ChunkInfo contains information of chunks.
type ChunkInfo struct {
	Type   string
	LocGid string
}

// ChunkHeader contains identification and information of chunk.
type ChunkHeader struct {
	Magic   [4]byte
	Type    []byte
	State   [1]byte
	Encoded bool
	_       [1]byte // padding
}

// ObjHeader contains identification and information of object.
type ObjHeader struct {
	Magic  [4]byte
	Name   []byte
	Size   int64
	Offset int64
}

// PartType is purpose of disk partition to be used.
type PartType string

const (
	// ArchivePart is a storage, which sometimes should be spined-down to reduce power consumption.
	ArchivePart PartType = "ArchivePart"

	// ServicePart is a storage, which is always spinning-up.
	ServicePart = "ServicePart"
)

// FsStatus contains space information of file system.
type FsStatus struct {
	All  uint64
	Used uint64
	Free uint64
}

// Lock contains locks for object and chunk respectively.
type Lock struct {
	Obj sync.RWMutex
	Chk sync.RWMutex
}

// PartInfo contains information of each partition.
type PartInfo struct {
	Size uint64
	Free uint64
	Used uint64
}

// PartGrpInfo contains
type PartGrpInfo struct {
	DiskSched uint
	NumOfPart uint
	PartInfo  map[string]PartInfo
}

// SubPartGroup contains information of partition groups related to be schedule.
type SubPartGroup struct {
	Cold PartGrpInfo
	Hot  PartGrpInfo
}

// Vol contains information about the volume.
type Vol struct {
	Name         string
	Dev          string
	MntPoint     string
	Size         uint64
	Free         uint64
	Used         uint64
	Speed        VolumeSpeed
	Status       Status
	NumOfPart    uint8
	SubPartGroup SubPartGroup

	ChunkSize int64
	ObjMap    map[string]ObjMap
	ChunkMap  map[string]ChunkMap
	Lock      Lock
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
		Dev:       dev,
		Status:    Prepared,
		ObjMap:    make(map[string]ObjMap),
		ChunkMap:  make(map[string]ChunkMap),
		NumOfPart: 1,
		SubPartGroup: SubPartGroup{
			Cold: PartGrpInfo{
				NumOfPart: 1,
			},
			Hot: PartGrpInfo{
				NumOfPart: 1,
			},
		},
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
	// speeds := [3]VolumeSpeed{High, Mid, Low}
	speeds := [2]VolumeSpeed{High, Mid}

	v.Speed = speeds[rand.Intn(len(speeds))]
}
