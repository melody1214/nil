package repository

import (
	"os"
	"os/exec"
	"testing"

	"github.com/chanyoung/nil/pkg/util/uuid"
)

func TestNewVol(t *testing.T) {
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

	v, err := NewVol(img, 5, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(v.MntPoint)
	defer v.Umount()

	t.Logf("%+v", *v)
	t.Logf("Volume usage: %d", v.Usage())
}
