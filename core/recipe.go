package albius

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/vanilla-os/oci"
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
		_, err = disk.NewPartition(name, fsType, start, end)
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

	for _, mnt := range recipe.Mountpoints {
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

		disk.Partitions[partInt-1].Mount(baseRoot + mnt.Target)
	}

	return nil
}

func (recipe *Recipe) setupFstabEntries() ([][]string, error) {
	fstabEntries := [][]string{}
	for _, mnt := range recipe.Mountpoints {
		entry := []string{}

		// Partition UUID
		uuid, err := GetUUIDByPath(mnt.Partition)
		if err != nil {
			return [][]string{}, err
		}

		// Partition fstype
		fstype, err := GetFilesystemByPath(mnt.Partition)
		if err != nil {
			return [][]string{}, err
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

		entry = append(entry, fmt.Sprintf("UUID=%s", uuid))
		entry = append(entry, mnt.Target)
		entry = append(entry, fstype)
		entry = append(entry, options)
		entry = append(entry, "0")
		entry = append(entry, "0")

		fstabEntries = append(fstabEntries, entry)
	}

	return fstabEntries, nil
}

func (recipe *Recipe) Install() error {
	switch recipe.Installation.Method {
	case UNSQUASHFS:
		err := Unsquashfs(recipe.Installation.Source, RootA, true)
		if err != nil {
			return err
		}
	case OCI:
		// TODO: Persistence directory
		// /etc  -> /persist/etc
		// /nix  -> /persist/nix
		// Won't use abroot-adapter, how to sync changes?
		err := oci.Write(recipe.Installation.Source, RootA)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported installation method '%s'", recipe.Installation.Method)
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
