package albius

import (
	"fmt"
	"os"
	"os/exec"
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
		makefsCmd := "mkfs.fat -F 16 %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Path))
	case FAT32:
		makefsCmd := "mkfs.fat -F 32 %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Path))
	case LINUX_SWAP:
		makefsCmd := "mkswap %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Path))
	case HFS, HFS_PLUS, UDF:
		return fmt.Errorf("Unsupported filesystem: %s", part.Filesystem)
	default:
		makefsCmd := "mkfs.%s %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Filesystem, part.Path))
	}

	if err != nil {
		return fmt.Errorf("Failed to make %s filesystem for %s: %s", part.Filesystem, part.Path, err)
	}

	return nil
}

// GenFstab
// UpdateInitramfs
