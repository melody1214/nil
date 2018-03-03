package store

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

// Speed represents a disk speed level.
type Speed int

const (
	// Low : 60 Mb/s < speed < 80 Mb/s
	Low Speed = iota
	// Mid : 80 Mb/s < speed < 100 Mb/s
	Mid
	// High : 100 Mb/s < speed
	High
)

// LV contains logical volume information
type LV struct {
	Name     string
	Dev      string
	MntPoint string
	Size     uint64
	Free     uint64
	Used     uint64
	Speed    Speed
}

// NewLV collects information about the logical volume with the given
// device path and returns a pointer of LV.
func NewLV(dev string) (lv *LV, err error) {
	// Get absolute device path.
	if dev, err = filepath.Abs(dev); err != nil {
		return nil, err
	}

	// Creates the LV with the given device path.
	lv = &LV{
		Dev: dev,
	}

	// Checks the given device path is valid.
	if err = lv.checkDevicePath(); err != nil {
		return nil, err
	}

	// Creates temporary mount point.
	lv.MntPoint = "tmp" + uuid.Gen()
	if err = os.Mkdir(lv.MntPoint, 0755); err != nil {
		return nil, err
	}
	defer os.RemoveAll(lv.MntPoint)

	if lv.MntPoint, err = filepath.Abs(lv.MntPoint); err != nil {
		return nil, err
	}

	// Mount to tmporary mount point.
	if err = lv.mount(); err != nil {
		return nil, err
	}
	defer lv.umount()

	// Update filesystem stats.
	if err = lv.updateStatFs(); err != nil {
		return nil, err
	}

	// TODO: Set the disk speed.
	lv.setSpeed()

	return
}

// checkDevice checks the device of logical volume exists.
func (l *LV) checkDevicePath() error {
	if _, err := os.Lstat(l.Dev); os.IsNotExist(err) {
		return fmt.Errorf("device with the given path is not exist")
	} else if err != nil {
		return fmt.Errorf("device with the given path is not valid: %v", err)
	}

	return nil
}

func (l *LV) mount() error {
	output, err := exec.Command("mount", l.Dev, l.MntPoint).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s: %v", output, err)
	}
	return err
}

func (l *LV) umount() error {
	output, err := exec.Command("umount", l.MntPoint).CombinedOutput()
	if err == nil {
		return nil
	}

	output, err = exec.Command("umount", l.Dev).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s: %v", output, err)
	}

	return err
}

func (l *LV) updateStatFs() error {
	fs := syscall.Statfs_t{}
	// if err := syscall.Statfs(l.Dev, &fs); err != nil {
	if err := syscall.Statfs(l.MntPoint, &fs); err != nil {
		return err
	}

	l.Size = fs.Blocks * uint64(fs.Bsize) / 1024 / 1024
	l.Free = fs.Bfree * uint64(fs.Bsize) / 1024 / 1024
	l.Used = l.Size - l.Free
	return nil
}

func (l *LV) usage() int {
	l.updateStatFs()

	if l.Size < 1 {
		return 0
	}

	return int((l.Used * 100) / l.Size)
}

func (l *LV) setSpeed() {
	speeds := [3]Speed{High, Mid, Low}

	l.Speed = speeds[rand.Intn(len(speeds))]
}
