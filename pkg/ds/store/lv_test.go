package store

import (
	"os"
	"os/exec"
	"testing"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

func TestNewLV(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("mount is only allowed for root user")
	}

	img := "image" + uuid.Gen()
	if _, err := exec.Command("dd", "bs=1M", "count=100", "if=/dev/zero", "of="+img).CombinedOutput(); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(img)

	if _, err := exec.Command("mkfs.xfs", img).CombinedOutput(); err != nil {
		t.Fatal(err)
	}

	lv, err := NewLV(img)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(lv.MntPoint)
	defer lv.umount()

	t.Logf("%+v", *lv)
	t.Logf("Volume usage: %d", lv.usage())
}
