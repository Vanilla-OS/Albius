package albius

import (
	"fmt"
	"regexp"
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
	// TODO: Handle crypto_LUKS filesystems
	mountCmd := "mount -m %s %s"

	err := RunCommand(fmt.Sprintf(mountCmd, part.Path, location))
	if err != nil {
		return fmt.Errorf("Failed to run mount command: %s", err)
	}

	return nil
}

func (part *Partition) UmountPartition() error {
	umountCmd := "umount %s"

	err := RunCommand(fmt.Sprintf(umountCmd, part.Path))
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
	resizePartCmd := "parted -s %s resizepart %s %d"

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
	if targetPathEnd >= 48 || targetPathEnd <= 57 {
	    target.Path = fmt.Sprintf("%sp%d", basePath, target.Number)
	} else {
	    target.Path = fmt.Sprintf("%s%d", basePath, target.Number)
    }
}
