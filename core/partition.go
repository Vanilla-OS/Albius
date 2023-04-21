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

func PartitionFromPath(path string) (*Partition, error) {
	newPartition := Partition{}
	newPartition.Path = path

	partInfoCmd := "parted -ms %s unit MiB print | sed '3p;d'"
	cmd := exec.Command("sh", "-c", fmt.Sprintf(partInfoCmd, path))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch partition %s: %s", path, err)
	}

	outSplits := strings.Split(string(output), ":")

	numberInt, err := strconv.Atoi(outSplits[0])
	if err != nil {
		return nil, err
	}
	newPartition.Number = numberInt

	newPartition.Start = outSplits[1]
	newPartition.End = outSplits[2]
	newPartition.Size = outSplits[3]
	newPartition.Filesystem = PartitionFs(outSplits[4])

	return &newPartition, nil
}

func (part *Partition) Mount(location string) error {
	// Pass mount operation to cryptsetup if it's a LUKS-encrypted partition
	luks, err := IsLuks(part)
	if err != nil {
		return err
	}
	if luks {
		partUUID, err := part.GetUUID()
		if err != nil {
			return err
		}
		return LuksOpen(part, fmt.Sprintf("luks-%s", partUUID))
	}

	mountCmd := "mount -m %s %s"

	// Check if device is already mounted at location
	checkPartCmd := "lsblk -n -o MOUNTPOINTS %s"
	cmd := exec.Command("sh", "-c", fmt.Sprintf(checkPartCmd, part.Path))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to locate partition %s: %s", part.Path, err)
	}
	if strings.Contains(string(output), location) {
		return nil
	}

	err = RunCommand(fmt.Sprintf(mountCmd, part.Path, location))
	if err != nil {
		return fmt.Errorf("Failed to run mount command: %s", err)
	}

	return nil
}

func (part *Partition) UmountPartition() error {
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
		return LuksClose(fmt.Sprintf("luks-%s", partUUID))
	}

	umountCmd := "umount %s"

	err = RunCommand(fmt.Sprintf(umountCmd, part.Path))
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
	lsblkCmd := "lsblk -n -o UUID %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, target.Path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get partition UUID: %s", err)
	}

	return string(output[:len(output)-1]), nil
}

func GetUUIDByPath(path string) (string, error) {
	lsblkCmd := "lsblk -n -o UUID %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get partition UUID: %s", err)
	}

	return string(output[:len(output)-1]), nil
}

func GetFilesystemByPath(path string) (string, error) {
	lsblkCmd := "lsblk -n -o FSTYPE %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get partition FSTYPE: %s", err)
	}

	return string(output[:len(output)-1]), nil
}
