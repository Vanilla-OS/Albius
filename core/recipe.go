package albius

import (
	"encoding/json"
	"fmt"
	"os"
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
	Method InstallationMethod
	Source string
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

	switch operation {
	case "label":
		label := DiskLabel(args[0].(string))
		err = disk.LabelDisk(label)
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
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
			err = MakeFs(part)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
		} else {
			_, err := disk.NewPartition(name, fsType, start, end)
			if err != nil {
				return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
			}
		}
	case "rm":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].RemovePartition()
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
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
	case "setflag":
		partNum, err := strconv.Atoi(args[0].(string))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
		err = disk.Partitions[partNum-1].SetPartitionFlag(args[1].(string), args[2].(bool))
		if err != nil {
			return fmt.Errorf("Failed to execute operation %s: %s", operation, err)
		}
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

	switch operation {
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
	case "timezone":
		tz := args[0].(string)
		err := SetTimezone(targetRoot, tz)
		if err != nil {
			return err
		}
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
	case "pkgremove":
		pkgRemovePath := args[0].(string)
		removeCmd := args[1].(string)
		err := RemovePackages(targetRoot, pkgRemovePath, removeCmd)
		if err != nil {
			return err
		}
	case "hostname":
		newHostname := args[0].(string)
		err := ChangeHostname(targetRoot, newHostname)
		if err != nil {
			return err
		}
	case "locale":
		localeCode := args[0].(string)
		err := SetLocale(targetRoot, localeCode)
		if err != nil {
			return err
		}
	case "swapon":
		partition := args[0].(string)
		err := Swapon(targetRoot, partition)
		if err != nil {
			return err
		}
	case "keyboard":
		layout := args[0].(string)
		model := args[1].(string)
		variant := args[2].(string)
		err := SetKeyboardLayout(targetRoot, layout, model, variant)
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
		err := OCISetup(recipe.Installation.Source, RootA, false)
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

	// Update Initramfs
	err = UpdateInitramfs(RootA)
	if err != nil {
		return fmt.Errorf("Failed to update initramfs: %s", err)
	}

	return nil
}
