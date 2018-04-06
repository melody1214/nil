package cmap

import (
	"os"
	"time"

	"github.com/chanyoung/nil/pkg/util/mlog"
)

const (
	lockFile = baseDir + "/" + "LOCK"
)

// lock locks updating map by LOCK file.
func lock() {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.Mkdir(baseDir, os.ModePerm)
	}

	for {
		f, err := os.OpenFile(lockFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, os.FileMode(0444))
		if err != nil {
			mlog.GetPackageLogger("pkg/cmap").Info("file lock waiting")

			time.Sleep(100 * time.Millisecond)
			continue
		}

		f.Close()
		break
	}
}

// unlock unlocks updating map by LOCK file.
func unlock() {
	os.RemoveAll(lockFile)
}
