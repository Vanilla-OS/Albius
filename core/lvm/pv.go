package lvm

import "fmt"

type Pv struct {
	Path, VgName, PvFmt string
	Attr                int
	Size, Free          float64
}

// PV attributes
const (
	PV_ATTR_MISSING     = 1 << iota
	PV_ATTR_EXPORTED    = 1 << iota
	PV_ATTR_DUPLICATE   = 1 << iota
	PV_ATTR_ALLOCATABLE = 1 << iota
	PV_ATTR_USED        = 1 << iota
)

func parsePvAttrs(attrStr string) (int, error) {
	attrVal := 0
	if attrStr[2] != '-' {
		attrVal += PV_ATTR_MISSING
	}
	if attrStr[1] != '-' {
		attrVal += PV_ATTR_EXPORTED
	}
	switch attrStr[0] {
	case 'd':
		attrVal += PV_ATTR_DUPLICATE
	case 'a':
		attrVal += PV_ATTR_ALLOCATABLE
	case 'u':
		attrVal += PV_ATTR_USED
	case '-':
	default:
		return -1, fmt.Errorf("invalid pv_attr: %s", attrStr)
	}

	return attrVal, nil
}

func FindPv(path string) (Pv, error) {
	lvm := NewLvm()
	pvs, err := lvm.Pvs(path)
	if err != nil {
		return Pv{}, fmt.Errorf("findPv: %v", err)
	}

	return pvs[0], nil
}

func (p *Pv) Remove() error {
	lvm := NewLvm()
	return lvm.Pvremove(p)
}

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
	return p.Attr&PV_ATTR_ALLOCATABLE > 0
}

func (p *Pv) IsUsed() bool {
	return p.Attr&PV_ATTR_USED > 0
}
