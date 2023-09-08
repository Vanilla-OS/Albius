package lvm

// #cgo LDFLAGS: -llvm2cmd
/*
#include "lvm_cgo.h"
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

func (l *Lvm) lvm2Run(command string, args ...interface{}) (string, error) {
	cmd := C.CString(fmt.Sprintf(command, args...))
	defer C.free(unsafe.Pointer(cmd))

	ret := C.lvm2_run(l._instance, cmd)
	if ret != ECMD_PROCESSED {
		return "", fmt.Errorf("command returned exit status %d", ret)
	}
	defer C.free(unsafe.Pointer(*C.log_output()))

	return C.GoString(*C.log_output()), nil
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
	_, err := l.lvm2Run("pvcreate -y %s", diskLabel)
	if err != nil {
		return fmt.Errorf("pvcreate: %v", err)
	}

	return nil
}

// pvs (list pvs)
func (l *Lvm) Pvs() ([]Pv, error) {
	output, err := l.lvm2Run("pvs --noheadings --units m --nosuffix --separator ,")
	if err != nil {
		return []Pv{}, fmt.Errorf("pvs: %v", err)
	}

	pvList := []Pv{}
	fmt.Println(output)
	pvs := strings.Split(output, "\n")
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
			return []Pv{}, fmt.Errorf("pvs: invalid pv_attr: %s", vals[3])
		}

		size, err := strconv.ParseFloat(vals[4], 64)
		if err != nil {
			return []Pv{}, fmt.Errorf("pvs: could not convert %s to float", vals[4])
		}
		free, err := strconv.ParseFloat(vals[5], 64)
		if err != nil {
			return []Pv{}, fmt.Errorf("pvs: could not convert %s to float", vals[5])
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

	return pvList, nil
}

// pvresize (resize pv)
func (l *Lvm) Pvresize(pv *Pv, setPvSize ...float64) error {
	setPvSizeOpt := ""
	if len(setPvSize) > 0 {
		setPvSizeOpt = fmt.Sprintf("--setphysicalvolumesize %fm", setPvSize[0])
	}

	_, err := l.lvm2Run("pvresize -y %s %s", setPvSizeOpt, pv.Path)
	if err != nil {
		return fmt.Errorf("pvresize: %v", err)
	}

	return nil
}

// pvmove (move phisical extents)
// TODO

// pvremove (make partition stop being a pv)
func (l *Lvm) Pvremove(pv interface{}) error {
	pvPath := ""
	switch pvar := pv.(type) {
	case string:
		pvPath += pvar
	case Pv:
		pvPath += pvar.Path
	default:
		return fmt.Errorf("invalid type for pv. Must be either a string with the PV's path or a PV struct")
	}

	_, err := l.lvm2Run("pvremove %s", pvPath)
	if err != nil {
		return fmt.Errorf("pvremove: %v", err)
	}

	return nil
}

// vgcreate (create vg)
func (l *Lvm) Vgcreate(name string, pvs ...interface{}) error {
	pvPaths := []string{}

	for _, pv := range pvs {
		switch pvar := pv.(type) {
		case string:
			pvPaths = append(pvPaths, pvar)
		case Pv:
			pvPaths = append(pvPaths, pvar.Path)
		default:
			return fmt.Errorf("invalid type for pv. Must be either a string with the PV's path or a PV struct")
		}
	}

	_, err := l.lvm2Run("vgcreate %s %s", name, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgcreate: %v", err)
	}

	return nil
}

// vgs (list vgs)
// vgchange (activate and deactivate vg)
// vgextend (add pv to vg)
// vgreduce (remove pv from vg)

// lvcreate (create lv)
// lvs (list lvs)
// lvrename (rename lv)
// lvresize (resize lv and fs)
// lvremove (remove lv)

func (p *Pv) IsMissing() bool {
	return p.Attr&PV_ATTR_MISSING > 0
}

func (p *Pv) IsExported() bool {
	return p.Attr&PV_ATTR_EXPORTED > 0
}

func (p *Pv) IsDuplicate() bool {
	return p.Attr&PV_ATTR_DUPLICATE > 0
}

func (p *Pv) IsAllocatable() bool {
	return p.Attr&PV_ATTR_ALLICATABLE > 0
}

func (p *Pv) IsUsed() bool {
	return p.Attr&PV_ATTR_USED > 0
}
