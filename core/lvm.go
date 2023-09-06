package albius

// #cgo LDFLAGS: -llvm2cmd
/*
#include <stdio.h>
#include <stdlib.h>
#include <lvm2cmd.h>

void lvm_log_capture_fn(int level, const char *file, int line,
						int dm_errno, const char *format)
{
	if (level != 4)
		return;

	printf("%s\n", format);
	return;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	ECMD_PROCESSED = iota + 1
	ENO_SUCH_CMD
	EINVALID_CMD_LINE
	EINIT_FAILED
	ECMD_FAILED
)

type Lvm struct {
	_instance unsafe.Pointer
}

func NewLvm() Lvm {
	instance := Lvm{
		C.lvm2_init(),
	}

	return instance
}

func (l *Lvm) Dispose() {
	C.lvm2_exit(l._instance)
}

// pvcreate (create pv)
func (l *Lvm) Pvcreate(diskLabel string) error {
	ret := C.lvm2_run(l._instance, C.CString("pvcreate -y "+diskLabel))
	if ret != ECMD_PROCESSED {
		return fmt.Errorf("pvcreate command returned exit status %d", ret)
	}

	return nil
}

// pvs (list pvs)
func (l *Lvm) Pvs() error {
	ret := C.lvm2_run(l._instance, C.CString("pvs"))
	if ret != ECMD_PROCESSED {
		return fmt.Errorf("pvcreate command returned exit status %d", ret)
	}

	return nil
}

// pvresize (resize pv)
// pvmove (move phisical extents)
// pvremove (make partition stop being a pv)

// vgcreate (create vg)
// vgs (list vgs)
// vgchange (activate and deactivate vg)
// vgextend (add pv to vg)
// vgreduce (remove pv from vg)

// lvcreate (create lv)
// lvs (list lvs)
// lvrename (rename lv)
// lvresize (resize lv and fs)
// lvremove (remove lv)
