package albius

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/vanilla-os/albius/core/lvm"
)

var (
	device  string
	lvmpart string
)

func TestMain(m *testing.M) {
	// Setup testing device
	// Create dummy image
	cmd := exec.Command("dd", "if=/dev/zero", "of=test.img", "count=102400")
	if err := cmd.Run(); err != nil {
		panic("error while creating testing device image: " + err.Error())
	}
	// Mount dummy image as loop device
	cmd = exec.Command("losetup", "--find", "--show", "test.img")
	cmd.Stderr = os.Stderr
	ret, err := cmd.Output()
	if err != nil {
		panic("error while mounting loop device: " + err.Error())
	}
	device = string(ret)
	device = device[:len(device)-1]

	// Create device label and add some partitions
	albiusDevice, err := LocateDisk(device)
	if err != nil {
		// Test: Show all devices
		out, err := exec.Command("lsblk", "-f").Output()
		if err != nil {
			panic("error listing block devices: " + err.Error())
		}
		fmt.Println(string(out))

		panic("error finding loop device: " + err.Error())
	}
	err = albiusDevice.LabelDisk(GPT)
	if err != nil {
		panic("error adding label to loop device: " + err.Error())
	}
	_, err = albiusDevice.NewPartition("", EXT4, 1, 25)
	if err != nil {
		panic("error creating partition A in loop device: " + err.Error())
	}
	_, err = albiusDevice.NewPartition("", EXT4, 26, -1)
	if err != nil {
		panic("error creating partition B in loop device: " + err.Error())
	}
	lvmpart = device + "p"

	// Run tests
	status := m.Run()

	// Remove testing device
	cmd = exec.Command("losetup", "-d", device)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("error while detaching testing device: " + err.Error())
	}
	err = os.Remove("test.img")
	if err != nil {
		panic("error while removing testing device image: " + err.Error())
	}

	// Cleanup
	os.Exit(status)
}

func TestPvcreate(t *testing.T) {
	err := lvm.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}
}

func TestPvs(t *testing.T) {
	pvs, err := lvm.Pvs()
	fmt.Printf(" -> Returned: %v\n", pvs)
	if err != nil {
		t.Error(err)
	}
}

func TestPvResize(t *testing.T) {
	pvs, err := lvm.Pvs()
	if err != nil {
		t.Error(err)
	}
	err = lvm.Pvresize(&pvs[0])
	if err != nil {
		t.Error(err)
	}
}

func TestPvShrink(t *testing.T) {
	pvs, err := lvm.Pvs()
	if err != nil {
		t.Error(err)
	}
	err = lvm.Pvresize(&pvs[0], 10.0)
	if err != nil {
		t.Error(err)
	}

	pvs, err = lvm.Pvs()
	fmt.Printf(" -> New size: %v\n", pvs)
	if err != nil {
		t.Error(err)
	}
}

func TestPvRemoveStr(t *testing.T) {
	err := lvm.Pvremove(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}
}

func TestPvRemoveStruct(t *testing.T) {
	// Recreate PV removed by previous test
	err := lvm.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}

	pvs, err := lvm.Pvs(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}

	err = lvm.Pvremove(&pvs[0])
	if err != nil {
		t.Error(err)
	}
}

func TestVgCreate(t *testing.T) {
	// Create two testing PVs
	err := lvm.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}
	err = lvm.Pvcreate(lvmpart + "2")
	if err != nil {
		t.Error(err)
	}

	// Pass one PV as struct and another as string
	pvs, err := lvm.Pvs(lvmpart + "1")
	if err != nil {
		t.Error(err)
	}

	err = lvm.Vgcreate("MyTestingVG", &pvs[0], lvmpart+"2")
	if err != nil {
		t.Error(err)
	}
}

func TestVgs(t *testing.T) {
	vgs, err := lvm.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestVgrename(t *testing.T) {
	// Retrieve Vg
	vgs, err := lvm.Vgs()
	if err != nil {
		t.Error(err)
	}

	err = vgs[0].Rename("MyTestingVG1")
	if err != nil {
		t.Error(err)
	}

	vgs, err = lvm.Vgs("MyTestingVG1")
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestVgReduce(t *testing.T) {
	// Retrieve Vg
	vgs, err := lvm.Vgs()
	if err != nil {
		t.Error(err)
	}

	err = vgs[0].Reduce(lvmpart + "2")
	if err != nil {
		t.Error(err)
	}

	// Retrieve Vg
	vgs, err = lvm.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestVgExtend(t *testing.T) {
	// Retrieve Vg
	vgs, err := lvm.Vgs()
	if err != nil {
		t.Error(err)
	}

	err = vgs[0].Extend(lvmpart + "2")
	if err != nil {
		t.Error(err)
	}

	// Retrieve Vg
	vgs, err = lvm.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Error(err)
	}
}

func TestLvCreate(t *testing.T) {
	err := lvm.Lvcreate("MyLv0", "MyTestingVG1", lvm.LV_TYPE_LINEAR, 30)
	if err != nil {
		t.Error(err)
	}
}

func TestLvs(t *testing.T) {
	lvs, err := lvm.Lvs()
	fmt.Printf(" -> Returned: %v\n", lvs)
	if err != nil {
		t.Error(err)
	}
}

func TestLvrename(t *testing.T) {
	// Retrieve Lv
	lv, err := lvm.FindLv("MyTestingVG1", "MyLv0")
	if err != nil {
		t.Error(err)
	}

	err = lv.Rename("MyLv1")
	if err != nil {
		t.Error(err)
	}

	lv, err = lvm.FindLv("MyTestingVG1", "MyLv1")
	fmt.Printf(" -> Returned: %v\n", lv)
	if err != nil {
		t.Error(err)
	}
}

func TestLvRemove(t *testing.T) {
	// Retrieve Lv
	lv, err := lvm.FindLv("MyTestingVG1", "MyLv1")
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
	vg, err := lvm.FindVg("MyTestingVG1")
	if err != nil {
		t.Error(err)
	}

	err = vg.Remove()
	if err != nil {
		t.Error(err)
	}
}
