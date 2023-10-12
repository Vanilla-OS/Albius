package lvm

import "fmt"

type Lv struct {
	Name, VgName, Pool string
	AttrVolType        int
	AttrPermissions    int
	AttrAllocPolicy    int
	AttrFixed          int
	AttrState          int
	AttrDevice         int
	AttrTargetType     int
	AttrBlocks         int
	AttrHealth         int
	AttrSkip           int
	Size               float64
}

type LVType string
type LVResizeMode int

const (
	LV_TYPE_LINEAR     = "linear"
	LV_TYPE_STRIPED    = "striped"
	LV_TYPE_SNAPSHOT   = "snapshot"
	LV_TYPE_RAID       = "raid"
	LV_TYPE_MIRROR     = "mirror"
	LV_TYPE_THIN       = "thin"
	LV_TYPE_THIN_POOL  = "thin-pool"
	LV_TYPE_VDO        = "vdo"
	LV_TYPE_VDO_POOL   = "vdo-pool"
	LV_TYPE_CACHE      = "cache"
	LV_TYPE_CACHE_POOL = "cache-pool"
	LV_TYPE_WRITECACHE = "writecache"
)

const (
	LV_RESIZE_EXTEND = iota
	LV_RESIZE_SHRINK = iota
)

const (
	LV_ATTR_VOL_TYPE_CACHE                    = 1 << iota
	LV_ATTR_VOL_TYPE_MIRRORED                 = 1 << iota
	LV_ATTR_VOL_TYPE_MIRRORED_NO_INITIAL_SYNC = 1 << iota
	LV_ATTR_VOL_TYPE_ORIGIN                   = 1 << iota
	LV_ATTR_VOL_TYPE_ORIGIN_MERGING_SNAPSHOT  = 1 << iota
	LV_ATTR_VOL_TYPE_RAID                     = 1 << iota
	LV_ATTR_VOL_TYPE_RAID_NO_INITIAL_SYNC     = 1 << iota
	LV_ATTR_VOL_TYPE_SNAPSHOT                 = 1 << iota
	LV_ATTR_VOL_TYPE_MERGING_SNAPSHOT         = 1 << iota
	LV_ATTR_VOL_TYPE_PVMOVE                   = 1 << iota
	LV_ATTR_VOL_TYPE_VIRTUAL                  = 1 << iota
	LV_ATTR_VOL_TYPE_IMAGE                    = 1 << iota
	LV_ATTR_VOL_TYPE_IMAGE_OUT_OF_SYNC        = 1 << iota
	LV_ATTR_VOL_TYPE_MIRROR_LOG_DEVICE        = 1 << iota
	LV_ATTR_VOL_TYPE_UNDER_CONVERSION         = 1 << iota
	LV_ATTR_VOL_TYPE_THIN_VOLUME              = 1 << iota
	LV_ATTR_VOL_TYPE_THIN_POOL                = 1 << iota
	LV_ATTR_VOL_TYPE_THIN_POOL_DATA           = 1 << iota
	LV_ATTR_VOL_TYPE_VDO_POOL                 = 1 << iota
	LV_ATTR_VOL_TYPE_VDO_POOL_DATA            = 1 << iota
	LV_ATTR_VOL_TYPE_METADATA                 = 1 << iota
)

const (
	LV_ATTR_PERMISSIONS_WRITEABLE              = 1 << iota
	LV_ATTR_PERMISSIONS_READONLY               = 1 << iota
	LV_ATTR_PERMISSIONS_READONLY_NON_RO_VOLUME = 1 << iota
)

const (
	LV_ATTR_ALLOC_POLICY_ANYWHERE   = 1 << iota
	LV_ATTR_ALLOC_POLICY_CONTIGUOUS = 1 << iota
	LV_ATTR_ALLOC_POLICY_INHERITED  = 1 << iota
	LV_ATTR_ALLOC_POLICY_CLING      = 1 << iota
	LV_ATTR_ALLOC_POLICY_NORMAL     = 1 << iota
)

const (
	LV_ATTR_FIXED_MINOR = 1 << iota
)

const (
	LV_ATTR_STATE_ACTIVE                                    = 1 << iota
	LV_ATTR_STATE_HISTORICAL                                = 1 << iota
	LV_ATTR_STATE_SUSPENDED                                 = 1 << iota
	LV_ATTR_STATE_INVALID_SNAPSHOT                          = 1 << iota
	LV_ATTR_STATE_INVALID_SUSPENDED_SNAPSHOT                = 1 << iota
	LV_ATTR_STATE_SNAPSHOT_MERGE_FAILED                     = 1 << iota
	LV_ATTR_STATE_SUSPENDED_SNAPSHOT_MERGE_FAILED           = 1 << iota
	LV_ATTR_STATE_MAPPED_DEVICE_PRESENT_WITHOUT_TABLES      = 1 << iota
	LV_ATTR_STATE_MAPPED_DEVICE_PRESENT_WITH_INACTIVE_TABLE = 1 << iota
	LV_ATTR_STATE_THIN_POOL_CHECK_NEEDED                    = 1 << iota
	LV_ATTR_STATE_SUSPENDED_THIN_POOL_CHECK_NEEDED          = 1 << iota
	LV_ATTR_STATE_UNKNOWN                                   = 1 << iota
)

const (
	LV_ATTR_DEVICE_OPEN    = 1 << iota
	LV_ATTR_DEVICE_UNKNOWN = 1 << iota
)

const (
	LV_ATTR_TARGET_TYPE_CACHE    = 1 << iota
	LV_ATTR_TARGET_TYPE_MIRROR   = 1 << iota
	LV_ATTR_TARGET_TYPE_RAID     = 1 << iota
	LV_ATTR_TARGET_TYPE_SNAPSHOT = 1 << iota
	LV_ATTR_TARGET_TYPE_THIN     = 1 << iota
	LV_ATTR_TARGET_TYPE_UNKNOWN  = 1 << iota
	LV_ATTR_TARGET_TYPE_VIRTUAL  = 1 << iota
)

const (
	LV_ATTR_BLOCKS_ARE_OVERWRITTEN_WITH_ZEROES_BEFORE_USE = 1 << iota
)

const (
	LV_ATTR_HEALTH_PARTIAL                      = 1 << iota
	LV_ATTR_HEALTH_UNKNOWN                      = 1 << iota
	LV_ATTR_HEALTH_RAID_REFRESH_NEEDED          = 1 << iota
	LV_ATTR_HEALTH_RAID_MISMATCHES_EXIST        = 1 << iota
	LV_ATTR_HEALTH_RAID_WRITEMOSTLY             = 1 << iota
	LV_ATTR_HEALTH_THIN_FAILED                  = 1 << iota
	LV_ATTR_HEALTH_THIN_POOL_OUT_OF_DATA_SPACE  = 1 << iota
	LV_ATTR_HEALTH_THIN_POOL_METADATA_READ_ONLY = 1 << iota
	LV_ATTR_HEALTH_WRITECACHE_ERROR             = 1 << iota
)

const (
	LV_ATTR_SKIP_ACTIVATION = 1 << iota
)

var (
	AttrVolTypeMap = map[byte]int{
		'C': LV_ATTR_VOL_TYPE_CACHE,
		'm': LV_ATTR_VOL_TYPE_MIRRORED,
		'M': LV_ATTR_VOL_TYPE_MIRRORED_NO_INITIAL_SYNC,
		'o': LV_ATTR_VOL_TYPE_ORIGIN,
		'O': LV_ATTR_VOL_TYPE_ORIGIN_MERGING_SNAPSHOT,
		'r': LV_ATTR_VOL_TYPE_RAID,
		'R': LV_ATTR_VOL_TYPE_RAID_NO_INITIAL_SYNC,
		's': LV_ATTR_VOL_TYPE_SNAPSHOT,
		'S': LV_ATTR_VOL_TYPE_MERGING_SNAPSHOT,
		'p': LV_ATTR_VOL_TYPE_PVMOVE,
		'v': LV_ATTR_VOL_TYPE_VIRTUAL,
		'i': LV_ATTR_VOL_TYPE_IMAGE,
		'I': LV_ATTR_VOL_TYPE_IMAGE_OUT_OF_SYNC,
		'l': LV_ATTR_VOL_TYPE_MIRROR_LOG_DEVICE,
		'c': LV_ATTR_VOL_TYPE_UNDER_CONVERSION,
		'V': LV_ATTR_VOL_TYPE_THIN_VOLUME,
		't': LV_ATTR_VOL_TYPE_THIN_POOL,
		'T': LV_ATTR_VOL_TYPE_THIN_POOL_DATA,
		'd': LV_ATTR_VOL_TYPE_VDO_POOL,
		'D': LV_ATTR_VOL_TYPE_VDO_POOL_DATA,
		'e': LV_ATTR_VOL_TYPE_METADATA,
		'-': 0,
	}

	AttrPermissionsMap = map[byte]int{
		'w': LV_ATTR_PERMISSIONS_WRITEABLE,
		'r': LV_ATTR_PERMISSIONS_READONLY,
		'R': LV_ATTR_PERMISSIONS_READONLY_NON_RO_VOLUME,
		'-': 0,
	}

	AttrAllocPolicyMap = map[byte]int{
		'a': LV_ATTR_ALLOC_POLICY_ANYWHERE,
		'c': LV_ATTR_ALLOC_POLICY_CONTIGUOUS,
		'i': LV_ATTR_ALLOC_POLICY_INHERITED,
		'l': LV_ATTR_ALLOC_POLICY_CLING,
		'n': LV_ATTR_ALLOC_POLICY_NORMAL,
		'-': 0,
	}

	AttrStateMap = map[byte]int{
		'a': LV_ATTR_STATE_ACTIVE,
		'h': LV_ATTR_STATE_HISTORICAL,
		's': LV_ATTR_STATE_SUSPENDED,
		'I': LV_ATTR_STATE_INVALID_SNAPSHOT,
		'S': LV_ATTR_STATE_INVALID_SUSPENDED_SNAPSHOT,
		'm': LV_ATTR_STATE_SNAPSHOT_MERGE_FAILED,
		'M': LV_ATTR_STATE_SUSPENDED_SNAPSHOT_MERGE_FAILED,
		'd': LV_ATTR_STATE_MAPPED_DEVICE_PRESENT_WITHOUT_TABLES,
		'i': LV_ATTR_STATE_MAPPED_DEVICE_PRESENT_WITH_INACTIVE_TABLE,
		'c': LV_ATTR_STATE_THIN_POOL_CHECK_NEEDED,
		'C': LV_ATTR_STATE_SUSPENDED_THIN_POOL_CHECK_NEEDED,
		'X': LV_ATTR_STATE_UNKNOWN,
		'-': 0,
	}

	AttrDeviceMap = map[byte]int{
		'o': LV_ATTR_DEVICE_OPEN,
		'X': LV_ATTR_DEVICE_UNKNOWN,
		'-': 0,
	}

	AttrTargetTypeMap = map[byte]int{
		'C': LV_ATTR_TARGET_TYPE_CACHE,
		'm': LV_ATTR_TARGET_TYPE_MIRROR,
		'r': LV_ATTR_TARGET_TYPE_RAID,
		's': LV_ATTR_TARGET_TYPE_SNAPSHOT,
		't': LV_ATTR_TARGET_TYPE_THIN,
		'u': LV_ATTR_TARGET_TYPE_UNKNOWN,
		'v': LV_ATTR_TARGET_TYPE_VIRTUAL,
		'-': 0,
	}

	AttrHealthMap = map[byte]int{
		'p': LV_ATTR_HEALTH_PARTIAL,
		'X': LV_ATTR_HEALTH_UNKNOWN,
		'r': LV_ATTR_HEALTH_RAID_REFRESH_NEEDED,
		'm': LV_ATTR_HEALTH_RAID_MISMATCHES_EXIST,
		'w': LV_ATTR_HEALTH_RAID_WRITEMOSTLY,
		'F': LV_ATTR_HEALTH_THIN_FAILED,
		'D': LV_ATTR_HEALTH_THIN_POOL_OUT_OF_DATA_SPACE,
		'M': LV_ATTR_HEALTH_THIN_POOL_METADATA_READ_ONLY,
		'E': LV_ATTR_HEALTH_WRITECACHE_ERROR,
		'-': 0,
	}
)

func parseLvAttrs(attrStr string) ([10]int, error) {
	attrVolType, ok := AttrVolTypeMap[attrStr[0]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrPermissions, ok := AttrPermissionsMap[attrStr[1]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrAllocPolicy, ok := AttrAllocPolicyMap[attrStr[2]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrFixed := 0
	if attrStr[3] == 'm' {
		attrFixed += LV_ATTR_FIXED_MINOR
	} else if attrStr[3] != '-' {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrState, ok := AttrStateMap[attrStr[4]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrDevice, ok := AttrDeviceMap[attrStr[5]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrTargetType, ok := AttrTargetTypeMap[attrStr[6]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrBlocks := 0
	if attrStr[7] == 'z' {
		attrFixed += LV_ATTR_BLOCKS_ARE_OVERWRITTEN_WITH_ZEROES_BEFORE_USE
	} else if attrStr[7] != '-' {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrHealth, ok := AttrHealthMap[attrStr[8]]
	if !ok {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}
	attrSkip := 0
	if attrStr[9] == 'k' {
		attrFixed += LV_ATTR_SKIP_ACTIVATION
	} else if attrStr[9] != '-' {
		return [10]int{}, fmt.Errorf("invalid lv_attr: %s", attrStr)
	}

	return [10]int{attrVolType, attrPermissions, attrAllocPolicy, attrFixed, attrState, attrDevice, attrTargetType, attrBlocks, attrHealth, attrSkip}, nil
}

// FindLv fetches an LVM logical volume by its name. You can pass the name
// either as a single string (as in `vg_name/lv_name`) or as two separate
// variables, where the first is the VG name and, the second, the LV name.
func FindLv(name string, lvName ...string) (Lv, error) {
	var fullName string
	if len(lvName) == 0 {
		fullName = name
	} else {
		fullName = name + "/" + lvName[0]
	}

	lvm := NewLvm()
	lvs, err := lvm.Lvs(fullName)
	if err != nil {
		return Lv{}, fmt.Errorf("findLv: %v", err)
	}

	return lvs[0], nil
}

func MakeThinPool(poolMetadata, pool interface{}) error {
	poolMetadataName, err := extractNameFromLv(poolMetadata)
	if err != nil {
		return err
	}
	poolName, err := extractNameFromLv(pool)
	if err != nil {
		return err
	}

	lvm := NewLvm()
	_, err = lvm.lvm2Run("lvconvert -y --type thin-pool --poolmetadata %s %s", poolMetadataName, poolName)
	if err != nil {
		return err
	}

	return nil
}

func (l *Lv) Rename(newName string) error {
	lvm := NewLvm()
	newLv, err := lvm.Lvrename(l.Name, newName, l.VgName)
	if err != nil {
		return err
	}
	*l = newLv

	return nil
}

func (l *Lv) Remove() error {
	lvm := NewLvm()
	return lvm.Lvremove(l)
}
