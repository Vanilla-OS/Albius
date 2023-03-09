package albius

import (
	"fmt"
	"regexp"
)

type Partition struct {
	Number                                   int
	Start, End, Size, Type, Filesystem, Path string
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

func (target *Partition) SetPartitionFlag(flag string, state int) error {
	setPartCmd := "parted -s %s set %s %s %s"

	var stateStr string
	if state == 0 {
		stateStr = "off"
	} else if state == 1 {
		stateStr = "on"
	} else {
        return fmt.Errorf("Invalid flag state: %d", state)
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
