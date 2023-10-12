package lvm

// #cgo LDFLAGS: -llvm2cmd
/*
#include "lvm_cgo.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// LVM command return codes
const (
	ECMD_PROCESSED    = iota + 1
	ENO_SUCH_CMD      = iota + 1
	EINVALID_CMD_LINE = iota + 1
	EINIT_FAILED      = iota + 1
	ECMD_FAILED       = iota + 1
)

type Lvm struct {
	_instance unsafe.Pointer
}

func (l *Lvm) lvm2Run(command string, args ...interface{}) (string, error) {
	cmd := C.CString(fmt.Sprintf(command, args...))
	defer C.free(unsafe.Pointer(cmd))

	ret := C.lvm2_run(l._instance, cmd)
	if ret != ECMD_PROCESSED {
		return "", fmt.Errorf("command returned exit status %d", ret)
	}

	output := ""
	for {
		logger := C.logger()
		if C.lvm_log_empty(logger) == 1 {
			break
		}
		entry := C.lvm_log_remove(&logger)
		C.set_logger(logger)
		entryStr := C.GoString((*C.char)(entry))
		C.free(entry)
		output += entryStr + "\n"
	}

	return output, nil
}

func NewLvm() Lvm {
	C.lvm2_log_fn((*[0]byte)(C.lvm_log_capture_fn))
	C.init_logger()

	instance := Lvm{
		C.lvm2_init(),
	}

	return instance
}

func (l *Lvm) Dispose() {
	C.free(unsafe.Pointer(C.logger()))
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
func (l *Lvm) Pvs(filter ...string) ([]Pv, error) {
	filterStr := ""
	if len(filter) > 0 {
		filterStr = strings.Join(filter, " ")
	}

	output, err := l.lvm2Run("pvs --noheadings --units m --nosuffix --separator , %s", filterStr)
	if err != nil {
		return []Pv{}, fmt.Errorf("pvs: %v", err)
	}

	pvList := []Pv{}
	pvs := strings.Split(output, "\n")
	for _, pv := range pvs {
		if !strings.HasPrefix(pv, "/") {
			continue
		}

		vals := strings.Split(pv, ",")
		attrVal, err := parsePvAttrs(vals[3])
		if err != nil {
			return []Pv{}, fmt.Errorf("pvs: %v", err)
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
func (l *Lvm) Pvresize(pv interface{}, setPvSize ...float64) error {
	pvPaths, err := extractPathsFromPvs(pv)
	if err != nil {
		return fmt.Errorf("pvresize: %v", err)
	}

	setPvSizeOpt := ""
	if len(setPvSize) > 0 {
		setPvSizeOpt = fmt.Sprintf("--setphysicalvolumesize %fm", setPvSize[0])
	}

	_, err = l.lvm2Run("pvresize -y %s %s", setPvSizeOpt, pvPaths[0])
	if err != nil {
		return fmt.Errorf("pvresize: %v", err)
	}

	return nil
}

// pvmove (move phisical extents)
// TODO

// pvremove (make partition stop being a pv)
func (l *Lvm) Pvremove(pv interface{}) error {
	pvPaths, err := extractPathsFromPvs(pv)
	if err != nil {
		return fmt.Errorf("pvremove: %v", err)
	}

	_, err = l.lvm2Run("pvremove -y %s", pvPaths[0])
	if err != nil {
		return fmt.Errorf("pvremove: %v", err)
	}

	return nil
}

// vgcreate (create vg)
func (l *Lvm) Vgcreate(name string, pvs ...interface{}) error {
	pvPaths, err := extractPathsFromPvs(pvs...)
	if err != nil {
		return fmt.Errorf("vgcreate: %v", err)
	}

	_, err = l.lvm2Run("vgcreate %s %s", name, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgcreate: %v", err)
	}

	return nil
}

// vgs (list vgs)
func (l *Lvm) Vgs(filter ...string) ([]Vg, error) {
	filterStr := ""
	if len(filter) > 0 {
		filterStr = strings.Join(filter, " ")
	}

	output, err := l.lvm2Run("vgs --noheadings --units m --nosuffix --separator , %s", filterStr)
	if err != nil {
		return []Vg{}, fmt.Errorf("vgs: %v", err)
	}

	vgList := []Vg{}
	vgs := strings.Split(output, "\n")
	for _, vg := range vgs {
		if vg == "" {
			continue
		}

		vals := strings.Split(vg, ",")
		attrVal, err := parseVgAttrs(vals[4])
		if err != nil {
			return []Vg{}, fmt.Errorf("vgs: %v", err)
		}
		size, err := strconv.ParseFloat(vals[5], 64)
		if err != nil {
			return []Vg{}, fmt.Errorf("vgs: could not convert %s to float", vals[4])
		}
		free, err := strconv.ParseFloat(vals[6], 64)
		if err != nil {
			return []Vg{}, fmt.Errorf("vgs: could not convert %s to float", vals[5])
		}

		// Filter Pvs matching Vg
		pvs, err := l.Pvs()
		if err != nil {
			return []Vg{}, fmt.Errorf("vgs: could not get list of pvs: %v", err)
		}
		matchedPvs := []Pv{}
		for _, pv := range pvs {
			if pv.VgName == vals[0] {
				matchedPvs = append(matchedPvs, pv)
			}
		}

		// Filter LVs matching Vg
		lvs, err := l.Lvs()
		if err != nil {
			return []Vg{}, fmt.Errorf("vgs: could not get list of lvs: %v", err)
		}
		matchedLvs := []Lv{}
		for _, lv := range lvs {
			if lv.VgName == vals[0] {
				matchedLvs = append(matchedLvs, lv)
			}
		}

		vgList = append(vgList, Vg{
			Name: vals[0],
			Pvs:  matchedPvs,
			Lvs:  matchedLvs,
			Attr: attrVal,
			Size: size,
			Free: free,
		})
	}

	return vgList, nil
}

// vgrename (rename vg)
func (l *Lvm) Vgrename(oldName, newName string) (Vg, error) {
	_, err := l.lvm2Run("vgrename %s %s", oldName, newName)
	if err != nil {
		return Vg{}, fmt.Errorf("vgrename: %v", err)
	}

	newVg, err := l.Vgs(newName)
	if err != nil {
		return Vg{}, fmt.Errorf("vgrename: %v", err)
	}

	return newVg[0], nil
}

// vgextend (add pv to vg)
func (l *Lvm) Vgextend(vg interface{}, pvs ...interface{}) error {
	if len(pvs) == 0 {
		return errors.New("vgextend: No PVs were provided")
	}

	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("vgextend: %v", err)
	}
	pvPaths, err := extractPathsFromPvs(pvs...)
	if err != nil {
		return fmt.Errorf("vgextend: %v", err)
	}

	_, err = l.lvm2Run("vgextend %s %s", vgName, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgextend: %v", err)
	}

	return nil
}

// vgreduce (remove pv from vg)
func (l *Lvm) Vgreduce(vg interface{}, pvs ...interface{}) error {
	if len(pvs) == 0 {
		return errors.New("vgreduce: No PVs were provided")
	}

	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("vgreduce: %v", err)
	}
	pvPaths, err := extractPathsFromPvs(pvs...)
	if err != nil {
		return fmt.Errorf("vgreduce: %v", err)
	}

	_, err = l.lvm2Run("vgreduce %s %s", vgName, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgreduce: %v", err)
	}

	return nil
}

// vgremove (remove vg)
func (l *Lvm) Vgremove(vg interface{}) error {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("vgremove: %v", err)
	}

	_, err = l.lvm2Run("vgremove -y %s", vgName)
	if err != nil {
		return fmt.Errorf("vgremove: %v", err)
	}

	return nil
}

// lvcreate (create lv)
func (l *Lvm) Lvcreate(name string, vg interface{}, lvType LVType, size float64) error {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("lvcreate: %v", err)
	}

	_, err = l.lvm2Run("lvcreate -y --type %s -L %.2fm %s -n %s", lvType, size, vgName, name)
	if err != nil {
		return fmt.Errorf("lvcreate: %v", err)
	}

	return nil
}

func (l *Lvm) LvThinCreate(name string, vg, pool interface{}, size float64) error {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("lvmThinCreate: %v", err)
	}

	poolName, err := extractNameFromPool(pool)
	if err != nil {
		return fmt.Errorf("lvmThinCreate: %v", err)
	}

	_, err = l.lvm2Run("lvcreate -y -n %s -V %.2fm --thinpool %s %s", name, size, poolName, vgName)
	if err != nil {
		return fmt.Errorf("lvmThinCreate: %v", err)
	}

	return nil
}

// lvs (list lvs)
func (l *Lvm) Lvs(filter ...string) ([]Lv, error) {
	filterStr := ""
	if len(filter) > 0 {
		filterStr = strings.Join(filter, " ")
	}

	output, err := l.lvm2Run("lvs --noheadings --units m --nosuffix --separator , %s", filterStr)
	if err != nil {
		return []Lv{}, fmt.Errorf("lvs: %v", err)
	}

	lvList := []Lv{}
	lvs := strings.Split(output, "\n")
	for _, lv := range lvs {
		if lv == "" {
			continue
		}

		vals := strings.Split(lv, ",")
		attrs, err := parseLvAttrs(vals[2])
		if err != nil {
			return []Lv{}, fmt.Errorf("lvs: %v", err)
		}

		size, err := strconv.ParseFloat(vals[3], 64)
		if err != nil {
			return []Lv{}, fmt.Errorf("lvs: could not convert %s to float", vals[5])
		}

		lvList = append(lvList, Lv{
			Name:            vals[0],
			VgName:          vals[1],
			Pool:            vals[4],
			AttrVolType:     attrs[0],
			AttrPermissions: attrs[1],
			AttrAllocPolicy: attrs[2],
			AttrFixed:       attrs[3],
			AttrState:       attrs[4],
			AttrDevice:      attrs[5],
			AttrTargetType:  attrs[6],
			AttrBlocks:      attrs[7],
			AttrHealth:      attrs[8],
			AttrSkip:        attrs[9],
			Size:            size,
		})
	}

	return lvList, nil
}

// lvrename (rename lv)
func (l *Lvm) Lvrename(oldName, newName string, vg interface{}) (Lv, error) {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return Lv{}, fmt.Errorf("lvrename: %v", err)
	}

	_, err = l.lvm2Run("lvrename %s %s %s", vgName, oldName, newName)
	if err != nil {
		return Lv{}, fmt.Errorf("lvrename: %v", err)
	}

	newLv, err := l.Lvs(vgName + "/" + newName)
	if err != nil {
		return Lv{}, fmt.Errorf("lvrename: %v", err)
	}

	return newLv[0], nil
}

// lvresize (resize lv and fs)
// TODO: Need to implement a function to resize filesystems first
// func (l *Lvm) Lvresize(lv interface{}, mode LVResizeMode, sizeOffset float64) error {
// 	return nil
// }

// lvremove (remove lv)
func (l *Lvm) Lvremove(lv interface{}) error {
	lvName, err := extractNameFromLv(lv)
	if err != nil {
		return fmt.Errorf("lvremove: %v", err)
	}

	_, err = l.lvm2Run("lvremove -y %s", lvName)
	if err != nil {
		return fmt.Errorf("lvrename: %v", err)
	}

	return nil
}

func extractPathsFromPvs(pvs ...interface{}) ([]string, error) {
	pvPaths := []string{}
	for _, pv := range pvs {
		switch pvar := pv.(type) {
		case string:
			pvPaths = append(pvPaths, pvar)
		case *Pv:
			pvPaths = append(pvPaths, pvar.Path)
		default:
			return nil, errors.New("invalid type for pv. Must be either a string with the PV's path or a pointer to a PV struct")
		}
	}

	return pvPaths, nil
}

func extractNameFromVg(vg interface{}) (string, error) {
	var vgName string
	switch vgvar := vg.(type) {
	case string:
		vgName = vgvar
	case *Vg:
		vgName = vgvar.Name
	default:
		return "", errors.New("invalid type for vg. Must be either a string with the VG's name or a pointer to a VG struct")
	}

	return vgName, nil
}

func extractNameFromLv(lv interface{}) (string, error) {
	var lvName string
	switch lvar := lv.(type) {
	case string:
		lvName = lvar
	case *Lv:
		lvName = lvar.VgName + "/" + lvar.Name
	default:
		return "", errors.New("invalid type for lv. Must be either a string with the LV's path ([group_name]/[lv_name]) or a pointer to a LV struct")
	}

	return lvName, nil
}

func extractNameFromPool(pool interface{}) (string, error) {
	var poolName string
	switch lvar := pool.(type) {
	case string:
		poolName = lvar
	case *Lv:
		poolName = lvar.Name
	default:
		return "", errors.New("invalid type for pool. Must be either a string with the pool's name or a pointer to a LV struct")
	}

	return poolName, nil
}
