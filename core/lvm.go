package albius

// #cgo LDFLAGS: -llvm2cmd
/*
#include <stdlib.h>
#include <string.h>
#include <lvm2cmd.h>

char *log_output = NULL;

void lvm_log_capture_fn(int level, const char *file, int line,
						int dm_errno, const char *format)
{
	if (level != 4)
		return;

	size_t log_len = strlen(format)+1;
	log_output = (char *)malloc(log_len * sizeof(char));
	memcpy(log_output, format, log_len);
	return;
}
*/
import "C"
import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

const (
	ECMD_PROCESSED    = iota + 1
	ENO_SUCH_CMD      = iota + 1
	EINVALID_CMD_LINE = iota + 1
	EINIT_FAILED      = iota + 1
	ECMD_FAILED       = iota + 1
)

const (
	PV_ATTR_MISSING     = 1 << iota
	PV_ATTR_EXPORTED    = 1 << iota
	PV_ATTR_DUPLICATE   = 1 << iota
	PV_ATTR_ALLICATABLE = 1 << iota
	PV_ATTR_USED        = 1 << iota
)

type Lvm struct {
	_instance unsafe.Pointer
}

type Pv struct {
	Path, VgName, PvFmt string
	Attr                int
	Size, Free          float64
}

func NewLvm() Lvm {
	C.lvm2_log_fn((*[0]byte)(C.lvm_log_capture_fn))

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
	command := C.CString("pvcreate -y " + diskLabel)
	ret := C.lvm2_run(l._instance, command)
	if ret != ECMD_PROCESSED {
		return fmt.Errorf("pvcreate command returned exit status %d", ret)
	}

	C.free(unsafe.Pointer(command))
	return nil
}

// pvs (list pvs)
func (l *Lvm) Pvs() ([]Pv, error) {
	command := C.CString("pvs --noheadings --units m --nosuffix --separator ,")
	ret := C.lvm2_run(l._instance, command)
	if ret != ECMD_PROCESSED {
		return []Pv{}, fmt.Errorf("pvcreate command returned exit status %d", ret)
	}

	pvList := []Pv{}
	pvs := strings.Split(C.GoString(C.log_output), "\n")
	for _, pv := range pvs {
		if pv == "" {
			continue
		}

		vals := strings.Split(pv, ",")

		attrVal := 0
		if vals[3][2] != '-' {
			attrVal += PV_ATTR_MISSING
		}
		if vals[3][1] != '-' {
			attrVal += PV_ATTR_EXPORTED
		}
		switch vals[3][0] {
		case 'd':
			attrVal += PV_ATTR_DUPLICATE
		case 'a':
			attrVal += PV_ATTR_ALLICATABLE
		case 'u':
			attrVal += PV_ATTR_USED
		case '-':
		default:
			return []Pv{}, fmt.Errorf("invalid pv_attr: %s", vals[3])
		}

		size, err := strconv.ParseFloat(vals[4], 64)
		if err != nil {
			return []Pv{}, fmt.Errorf("could not convert %s to float", vals[4])
		}
		free, err := strconv.ParseFloat(vals[5], 64)
		if err != nil {
			return []Pv{}, fmt.Errorf("coult not convert %s to float", vals[5])
		}

		pvList = append(pvList, Pv{
			Path:   vals[0],
			VgName: vals[1],
			PvFmt:  vals[2],
			Attr:   attrVal,
			Size:   size,
			Free:   free,
		})
	}

	C.free(unsafe.Pointer(C.log_output))
	C.free(unsafe.Pointer(command))

	return pvList, nil
}

// pvresize (resize pv)
func (l *Lvm) Pvresize(pv *Pv, setPvSize ...float64) error {
	setPvSizeOpt := ""
	if len(setPvSize) > 0 {
		setPvSizeOpt = fmt.Sprintf("--setphysicalvolumesize %fm", setPvSize[0])
	}
	command := C.CString(fmt.Sprintf("pvresize -y %s %s", setPvSizeOpt, pv.Path))
	ret := C.lvm2_run(l._instance, command)
	if ret != ECMD_PROCESSED {
		return fmt.Errorf("pvresize command returned exit status %d", ret)
	}

	return nil
}

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
