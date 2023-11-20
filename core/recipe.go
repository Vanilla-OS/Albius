package albius

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/vanilla-os/albius/core/lvm"
)

const (
	UNSQUASHFS = "unsquashfs"
	OCI        = "oci"
)

const (
	RootA = "/mnt/a"
	RootB = "/mnt/b"
)

type InstallationMethod string

type Recipe struct {
	Setup            []SetupStep
	Mountpoints      []Mountpoint
	Installation     Installation
	PostInstallation []PostStep
}

type SetupStep struct {
	Disk, Operation string
	Params          []interface{}
}

type Mountpoint struct {
	Partition, Target string
}

type Installation struct {
	Method        InstallationMethod
	Source        string
	InitramfsPre  []string
	InitramfsPost []string
}

type PostStep struct {
	Chroot    bool
	Operation string
	Params    []interface{}
}

func ReadRecipe(path string) (*Recipe, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe: %s", err)
	}

	var recipe Recipe
	err = json.Unmarshal(content, &recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe: %s", err)
	}

	return &recipe, nil
}

func runSetupOperation(diskLabel, operation string, args []interface{}) error {
	disk, err := LocateDisk(diskLabel)
	if err != nil {
		return err
	}

	/* !! ## Setup */
	switch operation {
	/* !! ### label
	 *
	 * Create a new partition table on the disk.
	 *
	 * **Accepts**:
	 * - *LabelType* (`string`): The partitioning scheme. Either `mbr` or `gpt`.
	 */
	case "label":
		label := DiskLabel(args[0].(string))
		err = disk.LabelDisk(label)
		if err != nil {
			return err
		}
	/* !! ### mkpart
	 *
	 * Create a new partition on the disk.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): The name for the partition.
	 * - *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`,
	 * `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*. If FsType
	 * is prefixed with `luks-` (e.g. `luks-btrfs`), the partition will be encrypted using LUKS2.
	 * - *Start* (`int`): The start position on disk for the new partition (in MiB).
	 * - *End* (`int`): The end position on disk for the new partition (in MiB), or -1 for
	 * using all the remaining space.
	 * - *LUKSPassword* (optional `string`): The password used to encrypt the partition. Only
	 * relevant if `FsType` is prefixed with `luks-`.
	 *
	 * \* = Not fully tested. Please create an issue if you encouter problems.
	 */
	case "mkpart":
		name := args[0].(string)
		fsType := PartitionFs(args[1].(string))
		start := int(args[2].(float64))
		end := int(args[3].(float64))
		if len(args) > 4 && strings.HasPrefix(string(fsType), "luks-") { // Encrypted partition
			luksPassword := args[4].(string)
			part, err := disk.NewPartition(name, "", start, end)
			if err != nil {
				return err
			}
			err = LuksFormat(part, luksPassword)
			if err != nil {
				return err
			}
			// lsblk seems to take a few milliseconds to update the partition's
			// UUID, so we loop until it gives us one
			uuid := ""
			for uuid == "" {
				uuid, _ = part.GetUUID()
			}
			err = LuksOpen(part, fmt.Sprintf("luks-%s", uuid), luksPassword)
			if err != nil {
				return err
			}
			part.Filesystem = PartitionFs(strings.TrimPrefix(string(fsType), "luks-"))
			err = LUKSMakeFs(part)
			if err != nil {
				return err
			}
			err = LUKSSetLabel(part, name)
			if err != nil {
				return err
			}
		} else { // Unencrypted partition
			_, err := disk.NewPartition(name, fsType, start, end)
			if err != nil {
				return err
			}
		}
	/* !! ### rm
	 *
	 * Delete a partition from the disk.
	 *
	 * **Accepts**:
	 * - *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 */
	case "rm":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return err
		}
		err = disk.Partitions[partNum-1].RemovePartition()
		if err != nil {
			return err
		}
	/* !! ### resizepart
	 *
	 * Resize a partition on disk.
	 *
	 * **Accepts**:
	 * - *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - *PartNewSize* (`int`): The new size in MiB for the partition.
	 */
	case "resizepart":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return err
		}
		partNewSize, err := strconv.Atoi(args[1].(string))
		if err != nil {
			return err
		}
		err = disk.Partitions[partNum-1].ResizePartition(partNewSize)
		if err != nil {
			return err
		}
	/* !! ### namepart
	 *
	 * Rename the specified partition.
	 *
	 * **Accepts**:
	 * - *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - *PartNewName* (`string`): The new name for the partition.
	 */
	case "namepart":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return err
		}
		partNewName := args[1].(string)
		if err != nil {
			return err
		}
		err = disk.Partitions[partNum-1].SetLabel(partNewName)
		if err != nil {
			return err
		}
		err = disk.Partitions[partNum-1].NamePartition(partNewName)
		if err != nil {
			return err
		}
	/* !! ### setflag
	 *
	 * Set the value of a partition flag, from the flags supported by parted.
	 * See parted(8) for the full list.
	 *
	 * **Accepts**:
	 * - *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - *FlagName* (`string`): The name of the flag.
	 * - *State* (`bool`): The value to apply to the flag. Either `true` or `false`.
	 */
	case "setflag":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return err
		}
		err = disk.Partitions[partNum-1].SetPartitionFlag(args[1].(string), args[2].(bool))
		if err != nil {
			return err
		}
	/* !! ### format
	 *
	 * Format an existing partition to a specified filesystem. This operation will destroy all data.
	 *
	 * **Accepts**:
	 * - *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
	 * - *Label* (optional `string`): An optional filesystem label. If not given, no label will be set.
	 */
	case "format":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return err
		}
		filesystem := args[1].(string)
		disk.Partitions[partNum-1].Filesystem = PartitionFs(filesystem)
		err = MakeFs(&disk.Partitions[partNum-1])
		if err != nil {
			return err
		}
		if len(args) == 3 {
			label := args[2].(string)
			err := disk.Partitions[partNum-1].SetLabel(label)
			if err != nil {
				return err
			}
		}
	/* !! ### luks-format
	 *
	 * Same as `format` but encrypts the partition with LUKS2.
	 *
	 * **Accepts**:
	 * - *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
	 * - *Password* (`string`): The password used to encrypt the partition.
	 * - *Label* (optional `string`): An optional filesystem label. If not given, no label will be set.
	 */
	case "luks-format":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return err
		}
		filesystem := args[1].(string)
		password := args[2].(string)
		part := disk.Partitions[partNum-1]
		part.Filesystem = PartitionFs(filesystem)
		err = LuksFormat(&part, password)
		if err != nil {
			return err
		}
		// lsblk seems to take a few milliseconds to update the partition's
		// UUID, so we loop until it gives us one
		uuid := ""
		for uuid == "" {
			uuid, _ = part.GetUUID()
		}
		err = LuksOpen(&part, fmt.Sprintf("luks-%s", uuid), password)
		if err != nil {
			return err
		}
		err = LUKSMakeFs(&part)
		if err != nil {
			return err
		}
		if len(args) == 4 {
			label := args[3].(string)
			err := LUKSSetLabel(&part, label)
			if err != nil {
				return err
			}
		}
	/* !! ### pvcreate
	 *
	 * Creates a new LVM physical volume from a partition.
	 *
	 * **Accepts**:
	 * - *Partition* (`string`): The partition to use as PV.
	 */
	case "pvcreate":
		part := args[0].(string)
		err := lvm.Pvcreate(part)
		if err != nil {
			return err
		}
	/* !! ### pvresize
	 *
	 * Resizes an LVM physical volume.
	 *
	 * **Accepts**:
	 * - *PV* (`string`): The physical volume path.
	 * - *Size* (optional `float`): The PV's desired size in MiB. If not provided, the PV will expand to the size of the underlying partition.
	 */
	case "pvresize":
		part := args[0].(string)
		var err error
		if len(args) > 1 {
			size := args[1].(float64)
			err = lvm.Pvresize(part, size)
		} else {
			err = lvm.Pvresize(part)
		}
		if err != nil {
			return err
		}
	/* !! ### pvremove
	 *
	 * Remove LVM labels from a partition.
	 *
	 * **Accepts**:
	 * - *PV* (`string`): The physical volume path.
	 */
	case "pvremove":
		part := args[0].(string)
		err := lvm.Pvremove(part)
		if err != nil {
			return err
		}
	/* !! ### vgcreate
	 *
	 * Creates a new LVM volume group.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): The VG name.
	 * - *PVs* (optional `[string]`): List containing paths for PVs to add to the newly created VG.
	 */
	case "vgcreate":
		name := args[0].(string)
		pvs := []string{}
		if len(args) > 1 {
			for _, pv := range args[1].([]interface{}) {
				pvs = append(pvs, pv.(string))
			}
		}
		pvList := make([]interface{}, len(pvs))
		for i, p := range pvs {
			pvList[i] = p
		}
		err := lvm.Vgcreate(name, pvList...)
		if err != nil {
			return err
		}
	/* !! ### vgrename
	 *
	 * Renames an LVM volume group.
	 *
	 * **Accepts**:
	 * - *OldName* (`string`): The VG's current name.
	 * - *NewName* (`string`): The VG's new name.
	 */
	case "vgrename":
		oldName := args[0].(string)
		newName := args[1].(string)
		_, err := lvm.Vgrename(oldName, newName)
		if err != nil {
			return err
		}
	/* !! ### vgextend
	 *
	 * Adds PVs to an LVM volume group.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): The target VG's name.
	 * - *PVs* (`[string]`): A list containing the paths of the PVs to be included.
	 */
	case "vgextend":
		name := args[0].(string)
		pvs := []string{}
		for _, pv := range args[1].([]interface{}) {
			pvs = append(pvs, pv.(string))
		}
		pvList := make([]interface{}, len(pvs))
		for i, p := range pvs {
			pvList[i] = p
		}
		err := lvm.Vgextend(name, pvList...)
		if err != nil {
			return err
		}
	/* !! ### vgreduce
	 *
	 * Removes PVs to an LVM volume group.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): The target VG's name.
	 * - *PVs* (`[string]`): A list containing the paths of the PVs to be removed.
	 */
	case "vgreduce":
		name := args[0].(string)
		pvs := []string{}
		for _, pv := range args[1].([]interface{}) {
			pvs = append(pvs, pv.(string))
		}
		pvList := make([]interface{}, len(pvs))
		for i, p := range pvs {
			pvList[i] = p
		}
		err := lvm.Vgreduce(name, pvList...)
		if err != nil {
			return err
		}
	/* !! ### vgremove
	 *
	 * Deletes LVM volume group.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): The volume group name.
	 */
	case "vgremove":
		name := args[0].(string)
		err := lvm.Vgremove(name)
		if err != nil {
			return err
		}
	/* !! ### lvcreate
	 *
	 * Create LVM logical volume.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): Logical volume name.
	 * - *VG* (`string`): Volume group name.
	 * - *Type* (`string`): Logical volume type. See lvcreate(8) for available types. If unsure, use `linear`.
	 * - *Size* (`float` or `string`): Logical volume size in MiB or a string containing a relative size (e.g. "100%FREE").
	 */
	case "lvcreate":
		name := args[0].(string)
		vg := args[1].(string)
		lvType := args[2].(string)
		vgSize := args[3]
		err := lvm.Lvcreate(name, vg, lvm.LVType(lvType), vgSize)
		if err != nil {
			return err
		}
	/* !! ### lvrename
	 *
	 * Renames an LVM logical volume.
	 *
	 * **Accepts**:
	 * - *OldName* (`string`): The LV's current name.
	 * - *NewName* (`string`): The LV's new name.
	 * - *VG* (`string`): Volume group the LV belongs to.
	 */
	case "lvrename":
		oldName := args[0].(string)
		newName := args[1].(string)
		vg := args[2].(string)
		_, err := lvm.Lvrename(oldName, newName, vg)
		if err != nil {
			return err
		}
	/* !! ### lvremove
	 *
	 * Deletes LVM logical volume.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): The logical volume name.
	 */
	case "lvremove":
		name := args[0].(string)
		err := lvm.Lvremove(name)
		if err != nil {
			return err
		}
	/* !! ### make-thin-pool
	 *
	 * Creates a new LVM thin pool from two LVs: one for metadata and another one for the data itself.
	 *
	 * **Accepts**:
	 * - *ThinDataLV* (`string`): The LV for storing data (in format `vg_name/lv_name`).
	 * - *ThinMetaLV* (`string`): The LV for storing pool metadata (in format `vg_name/lv_name`).
	 */
	case "make-thin-pool":
		thinDataLV := args[0].(string)
		thinMetaLV := args[1].(string)
		err := lvm.MakeThinPool(thinMetaLV, thinDataLV)
		if err != nil {
			return err
		}
	/* !! ### lvcreate-thin
	 *
	 * Same as `lvcreate`, but creates a thin LV instead.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): Thin logical volume name.
	 * - *VG* (`string`): Volume group name.
	 * - *Size* (`float`): Volume group size in MiB.
	 * - *Thinpool* (`string`): Name of the thin pool to create the LV from.
	 */
	case "lvcreate-thin":
		name := args[0].(string)
		vg := args[1].(string)
		vgSize := args[2].(float64)
		thinPool := args[3].(string)
		err := lvm.LvThinCreate(name, vg, thinPool, vgSize)
		if err != nil {
			return err
		}
	/* !! ### lvm-format
	 *
	 * Same as `format`, but formats an LVM logical volume.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): Thin logical volume name (in format `vg_name/lv_name`).
	 * - *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
	 * - *Label* (optional `string`): An optional filesystem label. If not given, no label will be set.
	 */
	case "lvm-format":
		name := args[0].(string)
		filesystem := args[1].(string)
		lv, err := lvm.FindLv(name)
		if err != nil {
			return err
		}
		dummyPart := Partition{
			Path:       "/dev/" + lv.VgName + "/" + lv.Name,
			Filesystem: PartitionFs(filesystem),
		}
		err = MakeFs(&dummyPart)
		if err != nil {
			return err
		}
		if len(args) == 3 {
			label := args[2].(string)
			err := dummyPart.SetLabel(label)
			if err != nil {
				return err
			}
		}
	/* !! ### lvm-luks-format
	 *
	 * Same as `luks-format`, but formats an LVM logical volume.
	 *
	 * **Accepts**:
	 * - *Name* (`string`): Thin logical volume name (in format `vg_name/lv_name`).
	 * - *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
	 * - *Password* (`string`): The password used to encrypt the volume.
	 * - *Label* (optional `string`): An optional filesystem label. If not given, no label will be set.
	 */
	case "lvm-luks-format":
		name := args[0].(string)
		filesystem := args[1].(string)
		password := args[2].(string)
		lv, err := lvm.FindLv(name)
		if err != nil {
			return err
		}
		dummyPart := Partition{
			Path:       "/dev/" + lv.VgName + "/" + lv.Name,
			Filesystem: PartitionFs(filesystem),
		}
		err = LuksFormat(&dummyPart, password)
		if err != nil {
			return err
		}
		// lsblk seems to take a few milliseconds to update the partition's
		// UUID, so we loop until it gives us one
		uuid := ""
		for uuid == "" {
			uuid, _ = dummyPart.GetUUID()
		}
		err = LuksOpen(&dummyPart, fmt.Sprintf("luks-%s", uuid), password)
		if err != nil {
			return err
		}
		err = LUKSMakeFs(&dummyPart)
		if err != nil {
			return err
		}
		if len(args) == 4 {
			label := args[3].(string)
			err := LUKSSetLabel(&dummyPart, label)
			if err != nil {
				return err
			}
		}
	/* !! --- */
	default:
		return fmt.Errorf("unrecognized operation %s", operation)
	}

	return nil
}

func (recipe *Recipe) RunSetup() error {
	for i, step := range recipe.Setup {
		fmt.Printf("Setup [%d/%d]: %s\n", i+1, len(recipe.Setup), step.Operation)
		err := runSetupOperation(step.Disk, step.Operation, step.Params)
		if err != nil {
			return fmt.Errorf("failed to run setup operation %s: %s", step.Operation, err)
		}
	}

	return nil
}

func runPostInstallOperation(chroot bool, operation string, args []interface{}) error {
	targetRoot := ""
	if chroot {
		targetRoot = RootA
	}

	/* !! ## Post-Installation */
	switch operation {
	/* !! ### adduser
	 *
	 * Create a new user.
	 *
	 * **Accepts**:
	 * - *Username* (`string`): The username of the new user.
	 * - *Fullname* (`string`): The full name (display name) of the new user.
	 * - *Groups* (`[string]`): A list of groups the new user belongs to (the new user is automatically part of its own group).
	 * - *Password* (optional `string`): The password for the user. If not provided, password login will be disabled.
	 */
	case "adduser":
		username := args[0].(string)
		fullname := args[1].(string)
		groups := []string{}
		for _, group := range args[2].([]interface{}) {
			groupStr := group.(string)
			groups = append(groups, groupStr)
		}
		var err error
		if len(args) == 4 {
			password := args[3].(string)
			err = AddUser(targetRoot, username, fullname, groups, password)
		} else {
			err = AddUser(targetRoot, username, fullname, groups)
		}
		if err != nil {
			return err
		}
	/* !! ### timezone
	 *
	 * Set the timezone.
	 *
	 * **Accepts**:
	 * - *TZ* (`string`): The timezone code (e.g. `America/Sao_Paulo`).
	 */
	case "timezone":
		tz := args[0].(string)
		err := SetTimezone(targetRoot, tz)
		if err != nil {
			return err
		}
	/* !! ### shell
	 *
	 * Run a shell command. This command accepts a variable number or parameters, where each parameter is a separate command to run.
	 *
	 * **Accepts**:
	 * - *Command(s)* (`...string`): The shell command(s) to execute.
	 */
	case "shell":
		for _, arg := range args {
			command := arg.(string)
			var err error
			if chroot {
				err = RunInChroot(targetRoot, command)
			} else {
				err = RunCommand(command)
			}
			if err != nil {
				return err
			}
		}
	/* !! ### pkgremove
	 *
	 * Given a file containing a list of packages, use the specified package manager to remove them.
	 *
	 * **Accepts**:
	 * - *PkgRemovePath* (`string`): The path containing the list of packages to remove.
	 * - *RemoveCmd* (`string`): The package manager command to remove packages (e.g. `apt remove`).
	 */
	case "pkgremove":
		pkgRemovePath := args[0].(string)
		removeCmd := args[1].(string)
		err := RemovePackages(targetRoot, pkgRemovePath, removeCmd)
		if err != nil {
			return err
		}
	/* !! ### hostname
	 *
	 * Set the system's hostname.
	 *
	 * **Accepts**:
	 * - *NewHostname* (`string`): The hostname to set.
	 */
	case "hostname":
		newHostname := args[0].(string)
		err := ChangeHostname(targetRoot, newHostname)
		if err != nil {
			return err
		}
	/* !! ### locale
	 *
	 * Set the system's locale, using `locale-gen` to generate the locale if not present.
	 *
	 * **Accepts**:
	 * - *LocaleCode* (`string`): The locale code to use. See `/etc/locale.gen` for the full list of locale codes.
	 */
	case "locale":
		localeCode := args[0].(string)
		err := SetLocale(targetRoot, localeCode)
		if err != nil {
			return err
		}
	/* !! ### swapon
	 *
	 * Use the provided partition as swap space.
	 *
	 * **Accepts**:
	 * - *Partition* (`string`): The partition to use as swap.
	 */
	case "swapon":
		partition := args[0].(string)
		err := Swapon(targetRoot, partition)
		if err != nil {
			return err
		}
	/* !! ### keyboard
	 *
	 * Set the system keyboard layout. See `keyboard(5)` for more details.
	 *
	 * **Accepts**:
	 * - *Layout* (`string`): The keyboard's layout (XKBLAYOUT).
	 * - *Model* (`string`): The keyboard's model (XKBMODEL).
	 * - *Variant* (`string`): The keyboard's variant (XKBVARIANT).
	 */
	case "keyboard":
		layout := args[0].(string)
		model := args[1].(string)
		variant := args[2].(string)
		err := SetKeyboardLayout(targetRoot, layout, model, variant)
		if err != nil {
			return err
		}
	/* !! ### grub-install
	 *
	 * Install GRUB to the specified partition.
	 *
	 * **Accepts**:
	 * - *BootDirectory* (`string`): The path for the boot dir (usually `/boot`).
	 * - *InstallDevice* (`string`): The disk where the boot partition is located.
	 * - *Target* (`string`): The target firmware. Either `bios` for legacy systems or `efi` for UEFI systems.
	 * - *EFIDevice* (optional `string`): Only required for EFI installations. The partition where the EFI is located.
	 */
	case "grub-install":
		bootDirectory := args[0].(string)
		installDevice := args[1].(string)
		target := args[2].(string)
		efiDevice := ""
		if len(args) > 3 {
			efiDevice = args[3].(string)
		}
		var grubTarget FirmwareType
		switch target {
		case "bios":
			grubTarget = BIOS
		case "efi":
			grubTarget = EFI
		default:
			return fmt.Errorf("failed to execute operation: %s: Unrecognized firmware type: '%s')", operation, target)
		}
		err := RunGrubInstall(targetRoot, bootDirectory, installDevice, grubTarget, efiDevice)
		if err != nil {
			return err
		}
	/* !! ### grub-default-config
	 *
	 * Write key-value pairs into `/etc/default/grub`. This command accepts a variable number of parameters, where each parameter represents a new item to add to the file.
	 *
	 * **Accepts**:
	 * - *KV(s)* (`...string`): The `KEY=value` pair(s) to add to the GRUB default file.
	 */
	case "grub-default-config":
		currentConfig, err := GetGrubConfig(targetRoot)
		if err != nil {
			return err
		}
		for _, arg := range args {
			kv := strings.SplitN(arg.(string), "=", 2)
			currentConfig[kv[0]] = kv[1]
		}
		err = WriteGrubConfig(targetRoot, currentConfig)
		if err != nil {
			return err
		}
	/* !! ### grub-add-script
	 *
	 * Add one or more script files into `/etc/default/grub.d`. This command accepts a variable number of parameters, where each parameter represents a new file to add to the directory.
	 *
	 * **Accepts**:
	 * - *ScriptPath(s)* (`...string`): The file path(s) for each script to add.
	 */
	case "grub-add-script":
		for _, arg := range args {
			scriptPath := arg.(string)
			err := AddGrubScript(targetRoot, scriptPath)
			if err != nil {
				return err
			}
		}
	/* !! ### grub-remove-script
	 *
	 * Remove one or more script files from `/etc/default/grub.d`. This command accepts a variable number of parameters, where each parameter represents a file to delete from the directory.
	 *
	 * **Accepts**:
	 * - *ScriptPath(s)* (`...string`): The file path(s) for each script to be removed.
	 */
	case "grub-remove-script":
		for _, arg := range args {
			scriptName := arg.(string)
			err := RemoveGrubScript(targetRoot, scriptName)
			if err != nil {
				return err
			}
		}
	/* !! ### grub-mkconfig
	 *
	 * Run the `grub-mkconfig` command to generate a new GRUB configuration into the specified output path.
	 *
	 * **Accepts**:
	 * - *OutputPath* (`string`): The target path for the generated config.
	 */
	case "grub-mkconfig":
		outputPath := args[0].(string)
		err := RunGrubMkconfig(targetRoot, outputPath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unrecognized operation %s", operation)
	}

	return nil
}

func (recipe *Recipe) RunPostInstall() error {
	for i, step := range recipe.PostInstallation {
		fmt.Printf("Post-installation [%d/%d]: %s\n", i+1, len(recipe.PostInstallation), step.Operation)
		err := runPostInstallOperation(step.Chroot, step.Operation, step.Params)
		if err != nil {
			return fmt.Errorf("failed to run post-install operation %s: %s", step.Operation, err)
		}
	}

	return nil
}

func (recipe *Recipe) SetupMountpoints() error {
	diskCache := map[string]*Disk{}
	rootAMounted := false

	/* We need to mount the partitions in order to prevent one mountpoint
	 * from overriding another.
	 * For example, if we mount /boot first in /mnt/a/boot and then mount / in
	 * /mnt/a, any files copied over to /mnt/a/boot will end up in the root
	 * partition.
	 */
	mount_depth := 0
	ordered_mountpoints := make([]*Mountpoint, 0)
	for len(ordered_mountpoints) < len(recipe.Mountpoints) {
		for i, mnt := range recipe.Mountpoints {
			cnt := strings.Count(mnt.Target, "/")
			if mnt.Target == "/" {
				cnt = 0
			}
			if cnt == mount_depth {
				ordered_mountpoints = append(ordered_mountpoints, &recipe.Mountpoints[i])
			}
		}
		mount_depth += 1
	}

	lvmExpr := regexp.MustCompile(`^/dev/(?P<vg>[\w-]+)/(?P<lv>[\w-]+)$`)

	for _, mnt := range ordered_mountpoints {
		baseRoot := RootA
		if mnt.Target == "/" && rootAMounted {
			baseRoot = RootB
		} else if mnt.Target == "/" && !rootAMounted {
			rootAMounted = true
		}

		// LVM partition
		if lvmExpr.MatchString(mnt.Partition) {
			lvmPartition := Partition{
				Number: -1,
				Path:   mnt.Partition,
			}
			err := lvmPartition.Mount(baseRoot + mnt.Target)
			if err != nil {
				return err
			}
			continue
		}

		// Regular partition
		diskName, partName := SeparateDiskPart(mnt.Partition)
		part, err := strconv.Atoi(partName)
		if err != nil {
			return err
		}

		disk, ok := diskCache[diskName]
		if !ok {
			diskPtr, err := LocateDisk(diskName)
			if err != nil {
				return err
			}
			diskCache[diskName] = diskPtr
			disk = diskCache[diskName]
		}

		err = disk.Partitions[part-1].Mount(baseRoot + mnt.Target)
		if err != nil {
			return err
		}
	}

	return nil
}

func (recipe *Recipe) setupFstabEntries() ([][]string, error) {
	fstabEntries := [][]string{}
	for _, mnt := range recipe.Mountpoints {
		entry := []string{}

		uuid, err := GetUUIDByPath(mnt.Partition)
		if err != nil {
			return [][]string{}, err
		}

		// Partition fstype
		fstype, err := GetFilesystemByPath(mnt.Partition)
		if err != nil {
			return [][]string{}, err
		}

		// If partition is LUKS-encrypted, use /dev/mapper/xxxx, otherwise
		// use the partition's UUID
		var fsName string
		luks, err := IsPathLuks(mnt.Partition)
		if err != nil {
			return [][]string{}, err
		}
		if luks {
			fsName = fmt.Sprintf("/dev/mapper/luks-%s", uuid)
			encryptedFstype, err := GetLUKSFilesystemByPath(mnt.Partition)
			if err != nil {
				return [][]string{}, err
			}
			fstype = encryptedFstype
		} else {
			fsName = fmt.Sprintf("UUID=%s", uuid)
		}

		// Partition options
		var options string
		switch mnt.Target {
		case "/boot/efi":
			options = "umask=0077"
		case "/boot":
			options = "noatime,errors=remount-ro"
		default:
			options = "defaults"
		}

		entry = append(entry, fsName, mnt.Target, fstype, options, "0", "0")
		fstabEntries = append(fstabEntries, entry)
	}

	return fstabEntries, nil
}

func (recipe *Recipe) setupCrypttabEntries() ([][]string, error) {
	crypttabEntries := [][]string{}
	for _, mnt := range recipe.Mountpoints {
		luks, err := IsPathLuks(mnt.Partition)
		if err != nil {
			return [][]string{}, err
		}
		if !luks {
			continue
		}

		entry := []string{}

		partUUID, err := GetUUIDByPath(mnt.Partition)
		if err != nil {
			return [][]string{}, err
		}

		entry = append(entry,
			fmt.Sprintf("luks-%s", partUUID), // target
			fmt.Sprintf("UUID=%s", partUUID), // device
			"none",                           // keyfile
			"luks,discard",                   // options
		)
		crypttabEntries = append(crypttabEntries, entry)
	}

	return crypttabEntries, nil
}

func (recipe *Recipe) Install() error {
	var err error
	switch recipe.Installation.Method {
	case UNSQUASHFS:
		err = Unsquashfs(recipe.Installation.Source, RootA, true)
	case OCI:
		err = OCISetup(recipe.Installation.Source, filepath.Join(RootA, "var"), RootA, false)
	default:
		return fmt.Errorf("unsupported installation method '%s'", recipe.Installation.Method)
	}
	if err != nil {
		return fmt.Errorf("failed to copy installation files: %s", err)
	}

	// Setup crypttab (if needed)
	crypttabEntries, err := recipe.setupCrypttabEntries()
	if err != nil {
		return fmt.Errorf("failed to generate crypttab entries: %s", err)
	}
	if len(crypttabEntries) > 0 {
		err = GenCrypttab(RootA, crypttabEntries)
		if err != nil {
			return fmt.Errorf("failed to generate crypttab: %s", err)
		}
	}

	// Setup fstab
	fstabEntries, err := recipe.setupFstabEntries()
	if err != nil {
		return fmt.Errorf("failed to generate fstab entries: %s", err)
	}
	err = GenFstab(RootA, fstabEntries)
	if err != nil {
		return fmt.Errorf("failed to generate fstab: %s", err)
	}

	// Initramfs pre-scripts
	for _, preCmd := range recipe.Installation.InitramfsPre {
		err := RunInChroot(RootA, preCmd)
		if err != nil {
			return fmt.Errorf("initramfs pre-script '%s' failed: %s", preCmd, err)
		}
	}

	// Update Initramfs
	err = UpdateInitramfs(RootA)
	if err != nil {
		return fmt.Errorf("failed to update initramfs: %s", err)
	}

	// Initramfs post-scripts
	for _, postCmd := range recipe.Installation.InitramfsPost {
		err := RunInChroot(RootA, postCmd)
		if err != nil {
			return fmt.Errorf("initramfs post-script '%s' failed: %s", postCmd, err)
		}
	}

	return nil
}
