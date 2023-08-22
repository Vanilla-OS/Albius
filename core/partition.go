package albius

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	BTRFS      = "btrfs"
	EXT2       = "ext2"
	EXT3       = "ext3"
	EXT4       = "ext4"
	FAT16      = "fat16"
	FAT32      = "fat32"
	HFS        = "hfs"
	HFS_PLUS   = "hfs+"
	LINUX_SWAP = "linux-swap"
	NTFS       = "ntfs"
	REISERFS   = "reiserfs"
	UDF        = "udf"
	XFS        = "xfs"
)

type PartitionFs string

type Partition struct {
	Number                       int
	Start, End, Size, Type, Path string
	Filesystem                   PartitionFs
}

func (part *Partition) Mount(location string) error {
	var mountPath string

	// If it's a LUKS-encrypted partition, open it first
	luks, err := IsLuks(part)
	if err != nil {
		return err
	}
	if luks {
		partUUID, err := part.GetUUID()
		if err != nil {
			return err
		}
		err = LuksTryOpen(part, fmt.Sprintf("luks-%s", partUUID), "")
		if err != nil {
			return err
		}

		mountPath, err = part.GetLUKSMapperPath()
		if err != nil {
			return err
		}
	} else {
		mountPath = part.Path
	}

	mountCmd := "mount -m %s %s"

	// Check if device is already mounted at location
	checkPartCmd := "lsblk -n -o MOUNTPOINTS %s"
	cmd := exec.Command("sh", "-c", fmt.Sprintf(checkPartCmd, mountPath))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to locate partition %s: %s", mountPath, err)
	}
	if strings.Contains(string(output), location) {
		return nil
	}

	err = RunCommand(fmt.Sprintf(mountCmd, mountPath, location))
	if err != nil {
		return fmt.Errorf("Failed to run mount command: %s", err)
	}

	return nil
}

func (part *Partition) IsMounted() (bool, error) {
	isMountedCmd := "mount | grep %s | wc -l"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(isMountedCmd, part.Path))
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("Failed to check if partition is mounted: %s", err)
	}

	mounts, err := strconv.Atoi(string(output))
	if err != nil {
		return false, fmt.Errorf("Failed to convert str to int: %s", err)
	}

	if mounts > 0 {
		return true, nil
	}

	return false, nil
}

func (part *Partition) UmountPartition() error {
	var mountTarget string

	// Check if partition is mounted first
	isMounted, err := part.IsMounted()
	if err != nil {
		return err
	}
	if !isMounted {
		return nil
	}

	// Pass unmount operation to cryptsetup if it's a LUKS-encrypted partition
	luks, err := IsLuks(part)
	if err != nil {
		return err
	}
	if luks {
		partUUID, err := part.GetUUID()
		if err != nil {
			return err
		}
		err = LuksClose(fmt.Sprintf("luks-%s", partUUID))
		if err != nil {
			return err
		}

		mountTarget, err = part.GetLUKSMapperPath()
		if err != nil {
			return err
		}
	} else {
		mountTarget = part.Path
	}

	umountCmd := "umount %s"

	err = RunCommand(fmt.Sprintf(umountCmd, mountTarget))
	if err != nil {
		return fmt.Errorf("Failed to run umount command: %s", err)
	}

	return nil
}

func UmountDirectory(dir string) error {
	umountCmd := "umount %s"

	err := RunCommand(fmt.Sprintf(umountCmd, dir))
	if err != nil {
		return fmt.Errorf("Failed to run umount command: %s", err)
	}

	return nil
}

func (target *Partition) RemovePartition() error {
	rmPartCmd := "parted -s %s rm %s"

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(target.Path)
	part := partExpr.FindString(target.Path)

	err := RunCommand(fmt.Sprintf(rmPartCmd, disk, part))
	if err != nil {
		return fmt.Errorf("Failed to remove partition: %s", err)
	}

	return nil
}

func (target *Partition) ResizePartition(newEnd int) error {
	resizePartCmd := "parted -s %s unit MiB resizepart %s %d"

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(target.Path)
	part := partExpr.FindString(target.Path)

	err := RunCommand(fmt.Sprintf(resizePartCmd, disk, part, newEnd))
	if err != nil {
		return fmt.Errorf("Failed to resize partition: %s", err)
	}

	return nil
}

func (target *Partition) NamePartition(name string) error {
	namePartCmd := "parted -s %s name %s %s"

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(target.Path)
	part := partExpr.FindString(target.Path)

	err := RunCommand(fmt.Sprintf(namePartCmd, disk, part, name))
	if err != nil {
		return fmt.Errorf("Failed to name partition: %s", err)
	}

	return nil
}

func (target *Partition) SetPartitionFlag(flag string, state bool) error {
	setPartCmd := "parted -s %s set %s %s %s"

	var stateStr string
	if !state {
		stateStr = "off"
	} else {
		stateStr = "on"
	}

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(target.Path)
	part := partExpr.FindString(target.Path)

	err := RunCommand(fmt.Sprintf(setPartCmd, disk, part, flag, stateStr))
	if err != nil {
		return fmt.Errorf("Failed to name partition: %s", err)
	}

	return nil
}

func (target *Partition) FillPath(basePath string) {
	targetPathEnd := basePath[len(basePath)-1]
	//                  "0"                    "9"
	if targetPathEnd >= 48 && targetPathEnd <= 57 {
		target.Path = fmt.Sprintf("%sp%d", basePath, target.Number)
	} else {
		target.Path = fmt.Sprintf("%s%d", basePath, target.Number)
	}
}

func (target *Partition) GetUUID() (string, error) {
	lsblkCmd := "lsblk -d -n -o UUID %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, target.Path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get partition UUID: %s", err)
	}

	return string(output[:len(output)-1]), nil
}

func GetUUIDByPath(path string) (string, error) {
	lsblkCmd := "lsblk -d -n -o UUID %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get partition UUID: %s", err)
	}

	return string(output[:len(output)-1]), nil
}

func GetFilesystemByPath(path string) (string, error) {
	lsblkCmd := "lsblk -d -n -o FSTYPE %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get partition FSTYPE: %s", err)
	}

	return string(output[:len(output)-1]), nil
}

func (part *Partition) GetLUKSMapperPath() (string, error) {
	// Assert part is a LUKS partition
	luks, err := IsLuks(part)
	if err != nil {
		return "", err
	}
	if !luks {
		return "", fmt.Errorf("Cannot get mapper path for %s. Partition is not LUKS-formatted", part.Path)
	}

	partUUID, err := part.GetUUID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/dev/mapper/luks-%s", partUUID), nil
}

func (part *Partition) SetLabel(label string) error {
	var labelCmd string
	switch part.Filesystem {
	case FAT16, FAT32:
		labelCmd = fmt.Sprintf("fatlabel %s %s", part.Path, label)
	case EXT2, EXT3, EXT4:
		labelCmd = fmt.Sprintf("e2label %s %s", part.Path, label)
	case BTRFS:
		labelCmd = fmt.Sprintf("btrfs filesystem label %s %s", part.Path, label)
	case REISERFS:
		labelCmd = fmt.Sprintf("reiserfstune –l %s %s", label, part.Path)
	case XFS:
		labelCmd = fmt.Sprintf("xfs_admin -L %s %s", label, part.Path)
	case LINUX_SWAP:
		return nil // There's no way to rename swap after it has been created
	case NTFS:
		labelCmd = fmt.Sprintf("ntfslabel %s %s", part.Path, label)
	default:
		return fmt.Errorf("Unsupported filesystem: %s", part.Filesystem)
	}

	err := RunCommand(labelCmd)
	if err != nil {
		return fmt.Errorf("Failed to label partition %s: %s", part.Path, err)
	}

	return nil
}

