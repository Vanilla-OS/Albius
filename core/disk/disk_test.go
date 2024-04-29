package disk

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"testing"

	luks "github.com/vanilla-os/albius/core/disk/luks"
)

var diskPath string

func TestMain(m *testing.M) {
	_, shouldSkip := os.LookupEnv("SKIP_ROOT_TESTS")
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	if currentUser.Username != "root" {
		if shouldSkip {
			log.Print("SKIP_ROOT_TESTS is set. Skipping test suite.")
			os.Exit(0)
		} else {
			log.Fatal("This test suite requires root privileges, which we do not have.\n\t\t    TIP: You can set the SKIP_ROOT_TESTS environment variable to ignore tests which require root access.")
		}
	}

	_, filename, _, _ := runtime.Caller(0)
	projRoot, _, _ := strings.Cut(filename, "core/")

	device, err := exec.Command(projRoot+"utils/create_test_device.sh", "-o", "test.img", "-s", "51200").Output()
	if err != nil {
		panic(err)
	}
	diskPath = strings.TrimSpace(string(device))

	status := m.Run()

	err = exec.Command(projRoot+"utils/remove_test_device.sh", diskPath, "test.img").Run()
	if err != nil {
		panic(err)
	}
	os.Exit(status)
}

func TestLocateDisk(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	if d.Path != diskPath {
		t.Errorf("Located incorrect disk: %v", d)
	}
}

func TestLabelDisk(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	err = d.LabelDisk(GPT)
	if err != nil {
		t.Error(err)
	}
}

func TestNewPartition(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	_, err = d.NewPartition("", EXT4, 1, 25)
	if err != nil {
		t.Error(err)
	}

	_, err = d.NewPartition("", EXT4, 26, -1)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPartition(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	if p := d.GetPartition(1); p == nil {
		t.Error(err)
	}
}

func TestLuksFormat(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	if err = luks.LuksFormat(&d.Partitions[0], "test"); err != nil {
		t.Error(err)
	}
}

func TestIsLuks(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	isLuks, err := luks.IsLuks(&d.Partitions[0])
	if err != nil {
		t.Error(err)
	}
	if !isLuks {
		t.Error("Failed to detect partition as LUKS-encrypted")
	}

	isLuks, err = luks.IsLuks(&d.Partitions[1])
	if err != nil {
		t.Error(err)
	}
	if isLuks {
		t.Error("Wrongly detected partition as LUKS-encrypted")
	}
}

func TestLuksOpen(t *testing.T) {
	d, err := LocateDisk(diskPath)
	if err != nil {
		t.Error(err)
	}

	err = luks.LuksOpen(&d.Partitions[0], "luks-test", "test")
	if err != nil {
		t.Error(err)
	}
}

func TestLuksClose(t *testing.T) {
	err := luks.LuksClose("luks-test")
	if err != nil {
		t.Error(err)
	}
}
