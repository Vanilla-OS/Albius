package lvm

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

// LVM command return codes
const (
	ECMD_PROCESSED    = iota + 1
	ENO_SUCH_CMD      = iota + 1
	EINVALID_CMD_LINE = iota + 1
	EINIT_FAILED      = iota + 1
	ECMD_FAILED       = iota + 1
)

func RunCommand(command string, args ...interface{}) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(command, args...))
	out, err := cmd.Output()

	exitErr, ok := err.(*exec.ExitError)
	if err != nil && ok {
		return strings.TrimSpace(string(out)), fmt.Errorf("lvm.RunCommand: \n%s", string(exitErr.Stderr))
	}

	return strings.TrimSpace(string(out)), err
}

// pvcreate (create pv)
func Pvcreate(diskLabel string) error {
	_, err := RunCommand("pvcreate -y %s", diskLabel)
	if err != nil {
		return fmt.Errorf("pvcreate: %v", err)
	}

	pvscanCache(diskLabel)

	return nil
}

// pvscanCache runs the command `pvscan --cache [pv]` on the specified volume
func pvscanCache(diskLabel string) error {
	_, err := RunCommand("pvscan --cache %s", diskLabel)
	if err != nil {
		return fmt.Errorf("pvscanCache: %v", err)
	}

	return nil
}

// pvs (list pvs)
func Pvs(filter ...string) ([]Pv, error) {
	filterStr := ""
	if len(filter) > 0 {
		filterStr = strings.Join(filter, " ")
	}

	output, err := RunCommand("pvs --noheadings --units m --nosuffix --separator , %s", filterStr)
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
func Pvresize(pv interface{}, setPvSize ...float64) error {
	pvPaths, err := extractPathsFromPvs(pv)
	if err != nil {
		return fmt.Errorf("pvresize: %v", err)
	}

	setPvSizeOpt := ""
	if len(setPvSize) > 0 {
		setPvSizeOpt = fmt.Sprintf("--setphysicalvolumesize %fm", setPvSize[0])
	}

	_, err = RunCommand("pvresize -y %s %s", setPvSizeOpt, pvPaths[0])
	if err != nil {
		return fmt.Errorf("pvresize: %v", err)
	}

	return nil
}

// pvmove (move phisical extents)
// TODO

// pvremove (make partition stop being a pv)
func Pvremove(pv interface{}) error {
	pvPaths, err := extractPathsFromPvs(pv)
	if err != nil {
		return fmt.Errorf("pvremove: %v", err)
	}

	_, err = RunCommand("pvremove -y %s", pvPaths[0])
	if err != nil {
		return fmt.Errorf("pvremove: %v", err)
	}

	return nil
}

// vgcreate (create vg)
func Vgcreate(name string, pvs ...interface{}) error {
	pvPaths, err := extractPathsFromPvs(pvs...)
	if err != nil {
		return fmt.Errorf("vgcreate: %v", err)
	}

	_, err = RunCommand("vgcreate %s %s", name, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgcreate: %v", err)
	}

	return nil
}

// vgs (list vgs)
func Vgs(filter ...string) ([]Vg, error) {
	filterStr := ""
	if len(filter) > 0 {
		filterStr = strings.Join(filter, " ")
	}

	output, err := RunCommand("vgs --noheadings --units m --nosuffix --separator , %s", filterStr)
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
		pvs, err := Pvs()
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
		lvs, err := Lvs()
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
func Vgrename(oldName, newName string) (Vg, error) {
	_, err := RunCommand("vgrename %s %s", oldName, newName)
	if err != nil {
		return Vg{}, fmt.Errorf("vgrename: %v", err)
	}

	newVg, err := Vgs(newName)
	if err != nil {
		return Vg{}, fmt.Errorf("vgrename: %v", err)
	}

	return newVg[0], nil
}

// vgextend (add pv to vg)
func Vgextend(vg interface{}, pvs ...interface{}) error {
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

	_, err = RunCommand("vgextend %s %s", vgName, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgextend: %v", err)
	}

	return nil
}

// vgreduce (remove pv from vg)
func Vgreduce(vg interface{}, pvs ...interface{}) error {
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

	_, err = RunCommand("vgreduce %s %s", vgName, strings.Join(pvPaths, " "))
	if err != nil {
		return fmt.Errorf("vgreduce: %v", err)
	}

	return nil
}

// vgremove (remove vg)
func Vgremove(vg interface{}) error {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("vgremove: %v", err)
	}

	_, err = RunCommand("vgremove -y %s", vgName)
	if err != nil {
		return fmt.Errorf("vgremove: %v", err)
	}

	return nil
}

// lvcreate (create lv)
func Lvcreate(name string, vg interface{}, lvType LVType, size interface{}) error {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("lvcreate: %v", err)
	}

	sizeStr := ""
	switch sizeVar := size.(type) {
	case string:
		sizeStr = "-l" + sizeVar
	case float64:
		sizeStr = fmt.Sprintf("-L %.2fm", sizeVar)
	case int:
		sizeStr = fmt.Sprintf("-L %dm", sizeVar)
	default:
		return fmt.Errorf("lvcreate: expected either string, int, or float64 for size, got %s", reflect.TypeOf(size))
	}

	_, err = RunCommand("lvcreate -y --type %s %s %s -n %s", lvType, sizeStr, vgName, name)
	if err != nil {
		return fmt.Errorf("lvcreate: %v", err)
	}

	return nil
}

func LvThinCreate(name string, vg, pool interface{}, size float64) error {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return fmt.Errorf("lvmThinCreate: %v", err)
	}

	poolName, err := extractNameFromPool(pool)
	if err != nil {
		return fmt.Errorf("lvmThinCreate: %v", err)
	}

	_, err = RunCommand("lvcreate -y -n %s -V %.2fm --thinpool %s %s", name, size, poolName, vgName)
	if err != nil {
		return fmt.Errorf("lvmThinCreate: %v", err)
	}

	return nil
}

// lvs (list lvs)
func Lvs(filter ...string) ([]Lv, error) {
	filterStr := ""
	if len(filter) > 0 {
		filterStr = strings.Join(filter, " ")
	}

	output, err := RunCommand("lvs --noheadings --units m --nosuffix --separator , %s", filterStr)
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
func Lvrename(oldName, newName string, vg interface{}) (Lv, error) {
	vgName, err := extractNameFromVg(vg)
	if err != nil {
		return Lv{}, fmt.Errorf("lvrename: %v", err)
	}

	_, err = RunCommand("lvrename %s %s %s", vgName, oldName, newName)
	if err != nil {
		return Lv{}, fmt.Errorf("lvrename: %v", err)
	}

	newLv, err := Lvs(vgName + "/" + newName)
	if err != nil {
		return Lv{}, fmt.Errorf("lvrename: %v", err)
	}

	return newLv[0], nil
}

// lvresize (resize lv and fs)
// TODO: Need to implement a function to resize filesystems first
// func  Lvresize(lv interface{}, mode LVResizeMode, sizeOffset float64) error {
// 	return nil
// }

// lvremove (remove lv)
func Lvremove(lv interface{}) error {
	lvName, err := extractNameFromLv(lv)
	if err != nil {
		return fmt.Errorf("lvremove: %v", err)
	}

	_, err = RunCommand("lvremove -y %s", lvName)
	if err != nil {
		return fmt.Errorf("lvremove: %v", err)
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
