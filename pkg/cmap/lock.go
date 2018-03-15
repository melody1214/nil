package cmap

import (
	"os"
	"time"

	"github.com/chanyoung/nil/pkg/util/mlog"
)

const (
	lockFile = baseDir + "/" + "LOCK"
)

// Lock locks updating map by LOCK file.
func Lock() {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.Mkdir(baseDir, os.ModePerm)
	}

	for {
		f, err := os.OpenFile(lockFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, os.FileMode(0444))
		if err != nil {
			mlog.GetLogger().WithField("pkg", "cmap").Info("file lock waiting")

			time.Sleep(100 * time.Millisecond)
			continue
		}

		f.Close()
		break
	}
}

// Unlock unlocks updating map by LOCK file.
func Unlock() {
	os.RemoveAll(lockFile)
}
