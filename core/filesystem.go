package albius

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Unsquashfs(filesystem, destination string, force bool) error {
	unsquashfsCmd := "unsquashfs%s -d %s %s"

	var forceFlag string
	if force {
		forceFlag = " -f"
	} else {
		forceFlag = ""
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf(unsquashfsCmd, forceFlag, destination, filesystem))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run unsquashfs: %s", err)
	}

	return nil
}

func MakeFs(part *Partition) error {
	var err error
	switch part.Filesystem {
	case FAT16:
		makefsCmd := "mkfs.fat -I -F 16 %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Path))
	case FAT32:
		makefsCmd := "mkfs.fat -I -F 32 %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Path))
	case LINUX_SWAP:
		makefsCmd := "mkswap -f %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Path))
	case HFS, HFS_PLUS, UDF:
		return fmt.Errorf("Unsupported filesystem: %s", part.Filesystem)
	default:
		makefsCmd := "mkfs.%s -f %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Filesystem, part.Path))
	}

	if err != nil {
		return fmt.Errorf("Failed to make %s filesystem for %s: %s", part.Filesystem, part.Path, err)
	}

	return nil
}

func GenFstab(targetRoot string, entries [][]string) error {
	fstabHeader := `# /etc/fstab: static file system information.
#
# Use 'blkid' to print the universally unique identifier for a
# device; this may be used with UUID= as a more robust way to name devices
# that works even if disks are added and removed. See fstab(5).
#
# <file system>  <mount point>  <type>  <options>  <dump>  <pass>`

	file, err := os.Create(fmt.Sprintf("%s/etc/fstab", targetRoot))
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(append([]byte(fstabHeader), '\n'))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fmtEntry := strings.Join(entry, " ")
		_, err = file.Write(append([]byte(fmtEntry), '\n'))
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateInitramfs(root string) error {
	updInitramfsCmd := "update-initramfs -c -k all"

	err := RunInChroot(root, updInitramfsCmd)
	if err != nil {
		return fmt.Errorf("Failed to run update-initramfs command: %s", err)
	}

	return nil
}

func RunInChroot(root, command string) error {
	cmd := exec.Command("chroot", root, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
