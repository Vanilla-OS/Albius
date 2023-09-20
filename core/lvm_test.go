package albius

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/vanilla-os/albius/core/lvm"
)

var (
	lvmInstance lvm.Lvm
	device      string
	lvmpart     string
)

func TestMain(m *testing.M) {
	// Create LVM wrapper instance
	lvmInstance = lvm.NewLvm()

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
	//Create device label and add some partitions
	albiusDevice, err := LocateDisk(device)
	if err != nil {
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
	lvmInstance.Dispose()
	os.Exit(status)
}

func TestPvcreate(t *testing.T) {
	err := lvmInstance.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvs(t *testing.T) {
	pvs, err := lvmInstance.Pvs()
	fmt.Printf(" -> Returned: %v\n", pvs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvResize(t *testing.T) {
	pvs, err := lvmInstance.Pvs()
	if err != nil {
		t.Fatal(err)
	}
	err = lvmInstance.Pvresize(&pvs[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvShrink(t *testing.T) {
	pvs, err := lvmInstance.Pvs()
	if err != nil {
		t.Fatal(err)
	}
	err = lvmInstance.Pvresize(&pvs[0], 10.0)
	if err != nil {
		t.Fatal(err)
	}

	pvs, err = lvmInstance.Pvs()
	fmt.Printf(" -> New size: %v\n", pvs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvRemoveStr(t *testing.T) {
	err := lvmInstance.Pvremove(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvRemoveStruct(t *testing.T) {
	// Recreate PV removed by previous test
	err := lvmInstance.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}

	pvs, err := lvmInstance.Pvs(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}

	err = lvmInstance.Pvremove(&pvs[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgCreate(t *testing.T) {
	// Create two testing PVs
	err := lvmInstance.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}
	err = lvmInstance.Pvcreate(lvmpart + "2")
	if err != nil {
		t.Fatal(err)
	}

	// Pass one PV as struct and another as string
	pvs, err := lvmInstance.Pvs(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}

	err = lvmInstance.Vgcreate("MyTestingVG", &pvs[0], lvmpart+"2")
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgs(t *testing.T) {
	vgs, err := lvmInstance.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgrename(t *testing.T) {
	// Retrieve Vg
	vgs, err := lvmInstance.Vgs()
	if err != nil {
		t.Fatal(err)
	}

	err = vgs[0].Rename("MyTestingVG1")
	if err != nil {
		t.Fatal(err)
	}

	vgs, err = lvmInstance.Vgs("MyTestingVG1")
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgReduce(t *testing.T) {
	// Retrieve Vg
	vgs, err := lvmInstance.Vgs()
	if err != nil {
		t.Fatal(err)
	}

	err = vgs[0].Reduce(lvmpart + "2")
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve Vg
	vgs, err = lvmInstance.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgExtend(t *testing.T) {
	// Retrieve Vg
	vgs, err := lvmInstance.Vgs()
	if err != nil {
		t.Fatal(err)
	}

	err = vgs[0].Extend(lvmpart + "2")
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve Vg
	vgs, err = lvmInstance.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLvCreate(t *testing.T) {
	err := lvmInstance.Lvcreate("MyLv0", "MyTestingVG1", lvm.LV_TYPE_LINEAR, 30)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLvs(t *testing.T) {
	lvs, err := lvmInstance.Lvs()
	fmt.Printf(" -> Returned: %v\n", lvs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLvrename(t *testing.T) {
	// Retrieve Lv
	lv, err := lvm.FindLv("MyTestingVG1", "MyLv0")
	if err != nil {
		t.Fatal(err)
	}

	err = lv.Rename("MyLv1")
	if err != nil {
		t.Fatal(err)
	}

	lv, err = lvm.FindLv("MyTestingVG1", "MyLv1")
	fmt.Printf(" -> Returned: %v\n", lv)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLvRemove(t *testing.T) {
	// Retrieve Lv
	lv, err := lvm.FindLv("MyTestingVG1", "MyLv1")
	if err != nil {
		t.Fatal(err)
	}

	err = lv.Remove()
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgRemove(t *testing.T) {
	// Retrieve Vg
	vg, err := lvm.FindVg("MyTestingVG1")
	if err != nil {
		t.Fatal(err)
	}

	err = vg.Remove()
	if err != nil {
		t.Fatal(err)
	}
}
