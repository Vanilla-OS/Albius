package albius

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
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

	// Check if partition is already mounted at location
	mountpoints, err := part.Mountpoints()
	if err != nil {
		return err
	}
	if slices.Contains(mountpoints, location) {
		return nil
	}

	mountCmd := "mount -m %s %s"
	err = RunCommand(fmt.Sprintf(mountCmd, mountPath, location))
	if err != nil {
		return fmt.Errorf("failed to run mount command: %s", err)
	}

	return nil
}

func (part *Partition) Mountpoints() ([]string, error) {
	mountpointsCmd := "lsblk -n -o MOUNTPOINTS %s"
	output, err := OutputCommand(fmt.Sprintf(mountpointsCmd, part.Path))
	if err != nil {
		return []string{}, fmt.Errorf("failed to list mountpoints for %s: %s", part.Path, err)
	}

	mounts := []string{}
	for _, mnt := range strings.Split(output, "\n") {
		if mnt != "" {
			mounts = append(mounts, mnt)
		}
	}

	return mounts, nil
}

func (part *Partition) IsMounted() (bool, error) {
	mountpoints, err := part.Mountpoints()
	if err != nil {
		return false, err
	}

	return len(mountpoints) > 0, nil
}

func (part *Partition) UnmountPartition() error {
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
		return fmt.Errorf("failed to run umount command: %s", err)
	}

	return nil
}

func UnmountDirectory(dir string) error {
	umountCmd := "umount %s"

	err := RunCommand(fmt.Sprintf(umountCmd, dir))
	if err != nil {
		return fmt.Errorf("failed to run umount command: %s", err)
	}

	return nil
}

func (target *Partition) RemovePartition() error {
	disk, part := SeparateDiskPart(target.Path)
	rmPartCmd := "parted -s %s rm %s"
	err := RunCommand(fmt.Sprintf(rmPartCmd, disk, part))
	if err != nil {
		return fmt.Errorf("failed to remove partition: %s", err)
	}

	return nil
}

func (target *Partition) ResizePartition(newEnd int) error {
	disk, part := SeparateDiskPart(target.Path)
	resizePartCmd := "parted -s %s unit MiB resizepart %s %d"
	err := RunCommand(fmt.Sprintf(resizePartCmd, disk, part, newEnd))
	if err != nil {
		return fmt.Errorf("failed to resize partition: %s", err)
	}

	return nil
}

func (target *Partition) NamePartition(name string) error {
	disk, part := SeparateDiskPart(target.Path)
	namePartCmd := "parted -s %s name %s %s"
	err := RunCommand(fmt.Sprintf(namePartCmd, disk, part, name))
	if err != nil {
		return fmt.Errorf("failed to name partition: %s", err)
	}

	return nil
}

func (target *Partition) SetPartitionFlag(flag string, state bool) error {
	stateStr := "off"
	if state {
		stateStr = "on"
	}

	disk, part := SeparateDiskPart(target.Path)
	setPartCmd := "parted -s %s set %s %s %s"
	err := RunCommand(fmt.Sprintf(setPartCmd, disk, part, flag, stateStr))
	if err != nil {
		return fmt.Errorf("failed to name partition: %s", err)
	}

	return nil
}

func (target *Partition) FillPath(basePath string) {
	targetPathEnd := basePath[len(basePath)-1]
	if targetPathEnd >= '0' && targetPathEnd <= '9' {
		target.Path = fmt.Sprintf("%sp%d", basePath, target.Number)
	} else {
		target.Path = fmt.Sprintf("%s%d", basePath, target.Number)
	}
}

func (target *Partition) GetUUID() (string, error) {
	lsblkCmd := "lsblk -d -n -o UUID %s"

	output, err := OutputCommand(fmt.Sprintf(lsblkCmd, target.Path))
	if err != nil {
		return "", fmt.Errorf("failed to get partition UUID: %s", err)
	}

	return output, nil
}

func GetUUIDByPath(path string) (string, error) {
	lsblkCmd := "lsblk -d -n -o UUID %s"

	output, err := OutputCommand(fmt.Sprintf(lsblkCmd, path))
	if err != nil {
		return "", fmt.Errorf("failed to get partition UUID: %s", err)
	}

	return output, nil
}

func GetFilesystemByPath(path string) (string, error) {
	lsblkCmd := "lsblk -d -n -o FSTYPE %s"

	output, err := OutputCommand(fmt.Sprintf(lsblkCmd, path))
	if err != nil {
		return "", fmt.Errorf("failed to get partition FSTYPE: %s", err)
	}

	return output, nil
}

func (part *Partition) GetLUKSMapperPath() (string, error) {
	// Assert part is a LUKS partition
	luks, err := IsLuks(part)
	if err != nil {
		return "", err
	}
	if !luks {
		return "", fmt.Errorf("cannot get mapper path for %s. Partition is not LUKS-formatted", part.Path)
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
		labelCmd = fmt.Sprintf("reiserfstune -l %s %s", label, part.Path)
	case XFS:
		labelCmd = fmt.Sprintf("xfs_admin -L %s %s", label, part.Path)
	case LINUX_SWAP:
		return nil // There's no way to rename swap after it has been created
	case NTFS:
		labelCmd = fmt.Sprintf("ntfslabel %s %s", part.Path, label)
	default:
		return fmt.Errorf("unsupported filesystem: %s", part.Filesystem)
	}

	err := RunCommand(labelCmd)
	if err != nil {
		return fmt.Errorf("failed to label partition %s: %s", part.Path, err)
	}

	return nil
}

// WaitUntilAvailable polls the specified partition until it is available.
// This is particularly useful to make sure a recently created or modified
// partition is recognized by the system.
func (part *Partition) WaitUntilAvailable() {
	for {
		_, err := os.Stat(part.Path)
		if !os.IsNotExist(err) {
			if uuid, err := part.GetUUID(); err != nil && uuid != "" {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}
