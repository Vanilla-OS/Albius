package lvm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"testing"
)

var lvmpart string

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

	device, err := exec.Command(projRoot+"utils/create_test_device.sh", "-o", "test.img", "-s", "51200", "-p", "\"\"", "ext4", "1", "25", "-p", "\"\"", "26", "100%").Output()
	if err != nil {
		panic(err)
	}
	deviceStr := strings.TrimSpace(string(device))
	lvmpart = deviceStr + "p"

	status := m.Run()

	err = exec.Command(projRoot+"utils/remove_test_device.sh", deviceStr, "test.img").Run()
	if err != nil {
		panic(err)
	}
	os.Exit(status)
}

func TestPvcreate(t *testing.T) {
	err := Pvcreate(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}
}

func TestPvs(t *testing.T) {
	pvs, err := Pvs()
	fmt.Printf(" -> Returned: %v\n", pvs)
	if err != nil {
		t.Error(err)
	}
}

func TestPvResize(t *testing.T) {
	pvs, err := Pvs()
	if err != nil {
		t.Error(err)
	}
	err = Pvresize(&pvs[0])
	if err != nil {
		t.Error(err)
	}
}

func TestPvShrink(t *testing.T) {
	pvs, err := Pvs()
	if err != nil {
		t.Error(err)
	}
	err = Pvresize(&pvs[0], 10.0)
	if err != nil {
		t.Error(err)
	}

	pvs, err = Pvs()
	fmt.Printf(" -> New size: %v\n", pvs)
	if err != nil {
		t.Error(err)
	}
}

func TestPvRemoveStr(t *testing.T) {
	err := Pvremove(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}
}

func TestPvRemoveStruct(t *testing.T) {
	// Recreate PV removed by previous test
	err := Pvcreate(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}

	pvs, err := Pvs(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}

	err = Pvremove(&pvs[0])
	if err != nil {
		t.Error(err)
	}
}

func TestVgCreate(t *testing.T) {
	// Create two testing PVs
	err := Pvcreate(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}
	err = Pvcreate(lvmpart + "2")
	if err != nil {
		t.Error(err)
	}

	// Pass one PV as struct and another as string
	pvs, err := Pvs(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}

	err = Vgcreate("MyTestingVG", &pvs[0], lvmpart+"2")
	if err != nil {
		t.Error(err)
	}
}

func TestVgs(t *testing.T) {
	vgs, err := Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestVgrename(t *testing.T) {
	// Retrieve Vg
	vgs, err := Vgs()
	if err != nil {
		t.Error(err)
	}

	err = vgs[0].Rename("MyTestingVG1")
	if err != nil {
		t.Error(err)
	}

	vgs, err = Vgs("MyTestingVG1")
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestVgReduce(t *testing.T) {
	// Retrieve Vg
	vgs, err := Vgs()
	if err != nil {
		t.Error(err)
	}

	err = vgs[0].Reduce(lvmpart + "2")
	if err != nil {
		t.Error(err)
	}

	// Retrieve Vg
	vgs, err = Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestVgExtend(t *testing.T) {
	// Retrieve Vg
	vgs, err := Vgs()
	if err != nil {
		t.Error(err)
	}

	err = vgs[0].Extend(lvmpart + "2")
	if err != nil {
		t.Error(err)
	}

	// Retrieve Vg
	vgs, err = Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestLvCreate(t *testing.T) {
	err := Lvcreate("MyLv0", "MyTestingVG1", LV_TYPE_LINEAR, 30)
	if err != nil {
		t.Error(err)
	}
}

func TestLvs(t *testing.T) {
	lvs, err := Lvs()
	fmt.Printf(" -> Returned: %v\n", lvs)
	if err != nil {
		t.Error(err)
	}
}

func TestLvrename(t *testing.T) {
	// Retrieve Lv
	lv, err := FindLv("MyTestingVG1", "MyLv0")
	if err != nil {
		t.Error(err)
	}

	err = lv.Rename("MyLv1")
	if err != nil {
		t.Error(err)
	}

	lv, err = FindLv("MyTestingVG1", "MyLv1")
	fmt.Printf(" -> Returned: %v\n", lv)
	if err != nil {
		t.Error(err)
	}
}

func TestLvRemove(t *testing.T) {
	// Retrieve Lv
	lv, err := FindLv("MyTestingVG1", "MyLv1")
	if err != nil {
		t.Error(err)
	}

	err = lv.Remove()
	if err != nil {
		t.Error(err)
	}
}

func TestVgRemove(t *testing.T) {
	// Retrieve Vg
	vg, err := FindVg("MyTestingVG1")
	if err != nil {
		t.Error(err)
	}

	err = vg.Remove()
	if err != nil {
		t.Error(err)
	}
}
