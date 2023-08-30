package albius

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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
		return nil, fmt.Errorf("Failed to read recipe: %s", err)
	}

	var recipe Recipe
	dec := json.NewDecoder(strings.NewReader(string(content)))
	dec.DisallowUnknownFields()
	dec.UseNumber()
	err = dec.Decode(&recipe)
	if err != nil {
		return nil, fmt.Errorf("Failed to read recipe: %s", err)
	}

	// Convert json.Number to int64
	for i := 0; i < len(recipe.Setup); i++ {
		step := &recipe.Setup[i]
		formattedParams := []interface{}{}
		for _, param := range step.Params {
			var dummy json.Number
			dummy = "1"
			if reflect.TypeOf(param) == reflect.TypeOf(dummy) {
				convertedParam, err := param.(json.Number).Int64()
				if err != nil {
					return nil, fmt.Errorf("Failed to convert recipe parameter: %s", err)
				}
				formattedParams = append(formattedParams, convertedParam)
			} else {
				formattedParams = append(formattedParams, param)
			}
		}
		step.Params = formattedParams
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
	 * Creates a new partition table on the disk.
	 *
	 * Accepts:
	 * - Label type (`string`): The partitioning scheme. Either `mbr` or `gpt`.
	 */
	case "label":
		label := DiskLabel(args[0].(string))
		err = disk.LabelDisk(label)
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
	/* !! ### mkpart
	 * Creates a new partition on the disk.
	 *
	 * Accepts:
	 * - Name (`string`): The name for the partition.
	 * - FsType (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*. If FsType is prefixed with `luks-` (e.g. `luks-btrfs`), the partition will be encrypted using LUKS2.
	 * - Start (`int`): The start position on disk for the new partition (in MiB).
	 * - End (`int`): The end position on disk for the new partition (in MiB), or -1 for using all the remaining space.
	 * - LUKSPassword (optional `string`): The password used to encrypt the partition. Only relevant if `FsType` is prefixed with `luks-`.
	 *
	 * \* = Not fully tested. Please create an issue if you encouter problems.
	 */
	case "mkpart":
		name := args[0].(string)
		fsType := PartitionFs(args[1].(string))
		start := args[2].(int64)
		end := args[3].(int64)
		if len(args) > 4 && strings.HasPrefix(string(fsType), "luks-") {
			luksPassword := args[4].(string)
			part, err := disk.NewPartition(name, "", start, end)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
			err = LuksFormat(part, luksPassword)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
			// lsblk seems to take a few milliseconds to update the partition's
			// UUID, so we loop until it gives us one
			uuid := ""
			for uuid == "" {
				uuid, err = part.GetUUID()
			}
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
			err = LuksOpen(part, fmt.Sprintf("luks-%s", uuid), luksPassword)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
			part.Filesystem = PartitionFs(strings.TrimPrefix(string(fsType), "luks-"))
			err = LUKSMakeFs(part)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
			err = LUKSSetLabel(part, name)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
		} else {
			_, err := disk.NewPartition(name, fsType, start, end)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
		}
	/* !! ### rm
	 * Deletes a partition from the disk.
	 *
	 * **Accepts**:
	 * - PartNum (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 */
	case "rm":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].RemovePartition()
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
	/* !! ### resizepart
	 * Resizes a partition on disk.
	 *
	 * **Accepts**:
	 * - PartNum (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - PartNewSize (`int`): The new size in MiB for the partition.
	 */
	case "resizepart":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		partNewSize, err := strconv.Atoi(args[1].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].ResizePartition(partNewSize)
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
	/* !! ### namepart
	 * Renames the specified partition.
	 *
	 * **Accepts**:
	 * - PartNum (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - PartNewName (`string`): The new name for the partition.
	 */
	case "namepart":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		partNewName := args[1].(string)
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].SetLabel(partNewName)
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].NamePartition(partNewName)
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
	/* !! ### setflag
	 * Sets the value of a partition flag, from the flags supported by parted. See parted(8) for the full list.
	 *
	 * **Accepts**:
	 * - PartNum (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - FlagName (`string`): The name of the flag.
	 * - State (`bool`): The value to apply to the flag. Either `true` or `false`.
	 */
	case "setflag":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].SetPartitionFlag(args[1].(string), args[2].(bool))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
	/* !! ### format
	 * Formats an existing partition to a specified filesystem. This operation will destroy all data.
	 *
	 * **Accepts**:
	 * - PartNum (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - FsType (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
	 */
	case "format":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		filesystem := args[1].(string)
		disk.Partitions[partNum-1].Filesystem = PartitionFs(filesystem)
		err = MakeFs(&disk.Partitions[partNum-1])
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
	/* !! ### luks-format
	 * Same as `format` but encrypts the partition with LUKS2.
	 *
	 * **Accepts**:
	 * - PartNum (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
	 * - FsType (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
	 * - Password (`string`): The password used to encrypt the partition.
	 */
	case "luks-format":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		filesystem := args[1].(string)
		password := args[2].(string)
		part := disk.Partitions[partNum-1]
		part.Filesystem = PartitionFs(filesystem)
		err = LuksFormat(&part, password)
		if err != nil {
			return err
		}
		err = LUKSMakeFs(&part)
		if err != nil {
			return err
		}
	/* !! --- */
	default:
		return fmt.Errorf("Unrecognized operation %s", operation)
	}

	return nil
}

func (recipe *Recipe) RunSetup() error {
	for _, step := range recipe.Setup {
		err := runSetupOperation(step.Disk, step.Operation, step.Params)
		if err != nil {
			return err
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
	 * Creates a new user.
	 *
	 * **Accepts**:
	 * - Username (`string`): The username of the new user.
	 * - Fullname (`string`): The full name (display name) of the new user.
	 * - Groups (`[string]`): A list of groups the new user belongs to (the new user is automatically part of its own group).
	 * - Password (optional `string`): The password for the user. If not proviced, password login will be disabled.
	 */
	case "adduser":
		username := args[0].(string)
		fullname := args[1].(string)
		groups := []string{}
		for _, group := range args[2].([]interface{}) {
			groupStr := group.(string)
			groups = append(groups, groupStr)
		}
		withPassword := false
		password := ""
		if len(args) == 4 {
			withPassword = true
			password = args[3].(string)
		}
		err := AddUser(targetRoot, username, fullname, groups, withPassword, password)
		if err != nil {
			return err
		}
	/* !! ### timezone
	 * Sets the timezone.
	 *
	 * **Accepts**:
	 * - TZ (`string`): The timezone code (e.g. `America/Sao_Paulo`).
	 */
	case "timezone":
		tz := args[0].(string)
		err := SetTimezone(targetRoot, tz)
		if err != nil {
			return err
		}
	/* !! ### shell
	 * Runs a shell command.
	 *
	 * **Accepts**:
	 * - Command(s) (`string` or `[string]`): The shell command(s) to execute.
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
	 * Given a file containing a list of packages, uses the specified package manager to remove them.
	 *
	 * **Accepts**:
	 * - PkgRemovePath (`string`): The path containing the list of packages to remove.
	 * - RemoveCmd (`string`): The package manager command to remove packages (e.g. `apt remove`).
	 */
	case "pkgremove":
		pkgRemovePath := args[0].(string)
		removeCmd := args[1].(string)
		err := RemovePackages(targetRoot, pkgRemovePath, removeCmd)
		if err != nil {
			return err
		}
	/* !! ### hostname
	 * Sets the system's hostname.
	 *
	 * **Accepts**:
	 * - NewHostname (`string`): The hostname to set.
	 */
	case "hostname":
		newHostname := args[0].(string)
		err := ChangeHostname(targetRoot, newHostname)
		if err != nil {
			return err
		}
	/* !! ### locale
	 * Sets the system's locale, using `locale-gen` to generate the locale if not present.
	 *
	 * **Accepts**:
	 * - LocaleCode (`string`): The locale code to use. See `/etc/locale.gen` for the full list of locale codes.
	 */
	case "locale":
		localeCode := args[0].(string)
		err := SetLocale(targetRoot, localeCode)
		if err != nil {
			return err
		}
	/* !! ### swapon
	 * Use the provided partition as swap space.
	 *
	 * **Accepts**:
	 * - Partition (`string`): The partition to use as swap.
	 */
	case "swapon":
		partition := args[0].(string)
		err := Swapon(targetRoot, partition)
		if err != nil {
			return err
		}
	/* !! ### keyboard
	 * Sets the system keyboard layout. See `keyboard(5)` for more details.
	 *
	 * **Accepts**:
	 * - Layout (`string`): The keyboard's layout (XKBLAYOUT).
	 * - Model (`string`): The keyboard's model (XKBMODEL).
	 * - Variant (`string`): The keyboard's variant (XKBVARIANT).
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
	 * Installs GRUB to the specified partition.
	 *
	 * **Accepts**:
	 * - BootDirectory (`string`): The path for the boot dir (usually `/boot`).
	 * - InstallDevice (`string`): The disk where the boot partition is located.
	 * - Target (`string`): The target firmware. Either `bios` for legacy systems or `efi` for UEFI systems.
	 */
	case "grub-install":
		bootDirectory := args[0].(string)
		installDevice := args[1].(string)
		target := args[2].(string)
		var grubTarget FirmwareType
		switch target {
		case "bios":
			grubTarget = BIOS
		case "efi":
			grubTarget = EFI
		default:
			return fmt.Errorf("Failed to execute operation: %s: Unrecognized firmware type: '%s')", operation, target)
		}
		err := RunGrubInstall(targetRoot, bootDirectory, installDevice, grubTarget)
		if err != nil {
			return err
		}
	/* !! ### grub-default-config
	 * TODO: Document
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
	 * TODO: Document
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
	 * TODO: Document
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
	 * TODO: Document
	 */
	case "grub-mkconfig":
		outputPath := args[0].(string)
		err := RunGrubMkconfig(targetRoot, outputPath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unrecognized operation %s", operation)
	}

	return nil
}

func (recipe *Recipe) RunPostInstall() error {
	for _, step := range recipe.PostInstallation {
		err := runPostInstallOperation(step.Chroot, step.Operation, step.Params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (recipe *Recipe) SetupMountpoints() error {
	diskCache := map[string]*Disk{}
	rootAMounted := false

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")

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

	for _, mnt := range ordered_mountpoints {
		diskName := diskExpr.FindString(mnt.Partition)
		part := partExpr.FindString(mnt.Partition)

		disk, ok := diskCache[diskName]
		if !ok {
			diskPtr, err := LocateDisk(diskName)
			if err != nil {
				return err
			}
			diskCache[diskName] = diskPtr
			disk = diskCache[diskName]
		}

		partInt, err := strconv.Atoi(part)
		if err != nil {
			return err
		}

		baseRoot := RootA
		if mnt.Target == "/" && rootAMounted {
			baseRoot = RootB
		} else if mnt.Target == "/" && !rootAMounted {
			rootAMounted = true
		}

		err = disk.Partitions[partInt-1].Mount(baseRoot + mnt.Target)
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

		entry = append(entry, fsName)
		entry = append(entry, mnt.Target)
		entry = append(entry, fstype)
		entry = append(entry, options)
		entry = append(entry, "0")
		entry = append(entry, "0")

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

		entry = append(entry, fmt.Sprintf("luks-%s", partUUID)) // target
		entry = append(entry, fmt.Sprintf("UUID=%s", partUUID)) // device
		entry = append(entry, "none")                           // keyfile
		entry = append(entry, "luks,discard")                   // options

		crypttabEntries = append(crypttabEntries, entry)
	}

	return crypttabEntries, nil
}

func (recipe *Recipe) Install() error {
	switch recipe.Installation.Method {
	case UNSQUASHFS:
		err := Unsquashfs(recipe.Installation.Source, RootA, true)
		if err != nil {
			return err
		}
	case OCI:
		err := OCISetup(recipe.Installation.Source, filepath.Join(RootA, "var"), RootA, false)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported installation method '%s'", recipe.Installation.Method)
	}

	// Setup crypttab (if needed)
	crypttabEntries, err := recipe.setupCrypttabEntries()
	if err != nil {
		return fmt.Errorf("Failed to generate crypttab entries: %s", err)
	}
	if len(crypttabEntries) > 0 {
		err = GenCrypttab(RootA, crypttabEntries)
		if err != nil {
			return fmt.Errorf("Failed to generate crypttab: %s", err)
		}
	}

	// Setup fstab
	fstabEntries, err := recipe.setupFstabEntries()
	if err != nil {
		return fmt.Errorf("Failed to generate fstab entries: %s", err)
	}
	err = GenFstab(RootA, fstabEntries)
	if err != nil {
		return fmt.Errorf("Failed to generate fstab: %s", err)
	}

	// Initramfs pre-scripts
	for _, preCmd := range recipe.Installation.InitramfsPre {
		err := RunInChroot(RootA, preCmd)
		if err != nil {
			return fmt.Errorf("Initramfs pre-script '%s' failed: %s", preCmd, err)
		}
	}

	// Update Initramfs
	err = UpdateInitramfs(RootA)
	if err != nil {
		return fmt.Errorf("Failed to update initramfs: %s", err)
	}

	// Initramfs post-scripts
	for _, postCmd := range recipe.Installation.InitramfsPost {
		err := RunInChroot(RootA, postCmd)
		if err != nil {
			return fmt.Errorf("Initramfs post-script '%s' failed: %s", postCmd, err)
		}
	}

	return nil
}
