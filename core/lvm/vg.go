package lvm

import "fmt"

type Vg struct {
	Name       string
	Pvs        []Pv
	Lvs        []Lv
	Attr       int
	Size, Free float64
}

// VG attributes
const (
	VG_ATTR_WRITABLE   = 1 << iota
	VG_ATTR_READONLY   = 1 << iota
	VG_ATTR_RESIZABLE  = 1 << iota
	VG_ATTR_EXPORTED   = 1 << iota
	VG_ATTR_PARTIAL    = 1 << iota
	VG_ATTR_CONTIGUOUS = 1 << iota
	VG_ATTR_CLING      = 1 << iota
	VG_ATTR_NORMAL     = 1 << iota
	VG_ATTR_ANYWHERE   = 1 << iota
	VG_ATTR_CLUSTERED  = 1 << iota
	VG_ATTR_SHARED     = 1 << iota
)

func ParseVgAttrs(attrStr string) (int, error) {
	attrVal := 0
	switch attrStr[5] {
	case 'c':
		attrVal += VG_ATTR_CLUSTERED
	case 's':
		attrVal += VG_ATTR_SHARED
	case '-':
	default:
		return -1, fmt.Errorf("invalid vg_attr: %s", attrStr)
	}
	switch attrStr[4] {
	case 'c':
		attrVal += VG_ATTR_CONTIGUOUS
	case 'l':
		attrVal += VG_ATTR_CLING
	case 'n':
		attrVal += VG_ATTR_NORMAL
	case 'a':
		attrVal += VG_ATTR_ANYWHERE
	default:
		return -1, fmt.Errorf("invalid vg_attr: %s", attrStr)
	}
	if attrStr[3] != '-' {
		attrVal += VG_ATTR_PARTIAL
	}
	if attrStr[2] != '-' {
		attrVal += VG_ATTR_EXPORTED
	}
	if attrStr[1] != '-' {
		attrVal += VG_ATTR_RESIZABLE
	}
	switch attrStr[0] {
	case 'w':
		attrVal += VG_ATTR_WRITABLE
	case 'r':
		attrVal += VG_ATTR_READONLY
	default:
		return -1, fmt.Errorf("invalid vg_attr: %s", attrStr)
	}

	return attrVal, nil
}

func FindVg(path string) (Vg, error) {
	lvm := NewLvm()
	vgs, err := lvm.Vgs(path)
	if err != nil {
		return Vg{}, fmt.Errorf("findVg: %v", err)
	}
	lvm.Dispose()

	return vgs[0], nil
}

func (v *Vg) IsWritable() bool {
	return v.Attr&VG_ATTR_WRITABLE > 0
}

func (v *Vg) IsReadonly() bool {
	return v.Attr&VG_ATTR_READONLY > 0
}

func (v *Vg) IsResizable() bool {
	return v.Attr&VG_ATTR_RESIZABLE > 0
}

func (v *Vg) IsExported() bool {
	return v.Attr&VG_ATTR_EXPORTED > 0
}

func (v *Vg) IsPartial() bool {
	return v.Attr&VG_ATTR_PARTIAL > 0
}

func (v *Vg) IsContiguous() bool {
	return v.Attr&VG_ATTR_CONTIGUOUS > 0
}

func (v *Vg) IsCling() bool {
	return v.Attr&VG_ATTR_CLING > 0
}

func (v *Vg) IsNormal() bool {
	return v.Attr&VG_ATTR_NORMAL > 0
}

func (v *Vg) IsAnywhere() bool {
	return v.Attr&VG_ATTR_ANYWHERE > 0
}

func (v *Vg) IsClustered() bool {
	return v.Attr&VG_ATTR_CLUSTERED > 0
}

func (v *Vg) IsShared() bool {
	return v.Attr&VG_ATTR_SHARED > 0
}
