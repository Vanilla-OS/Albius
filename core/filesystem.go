package albius

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/prometheus"
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
	case EXT2, EXT3, EXT4:
		makefsCmd := "mkfs.%s -F %s"
		err = RunCommand(fmt.Sprintf(makefsCmd, part.Filesystem, part.Path))
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

func OCISetup(imageSource, destination string, verbose bool) error {
	if reexec.Init() { // needed for subprocesses
		return fmt.Errorf("Failed to initialize reexec")
	}

	pmt, err := prometheus.NewPrometheus("storage", "overlay")
	if err != nil {
		return fmt.Errorf("Failed to create Prometheus instance: %s", err)
	}

	storedImageName := strings.ReplaceAll(imageSource, "/", "-")
	manifest, err := pmt.PullImage(imageSource, storedImageName)
	if err != nil {
		return fmt.Errorf("Failed to pull OCI image: %s", err)
	}

	fmt.Printf("Image pulled with digest %s\n", manifest.Config.Digest)

	image, err := pmt.GetImageByDigest(manifest.Config.Digest)
	if err != nil {
		return fmt.Errorf("Failed to get image from digest: %s", err)
	}

	mountPoint, err := pmt.MountImage(image.TopLayer)
	if err != nil {
		return fmt.Errorf("Failed to mount image at %s: %s", image.TopLayer, err)
	}

	fmt.Printf("Image mounted at %s\n", mountPoint)

	// Rsync image into destination
	fmt.Printf("Copying image to %s\n", destination)
	var verboseFlag string
	if verbose {
		verboseFlag = "v"
	} else {
		verboseFlag = ""
	}
	err = RunCommand(fmt.Sprintf("rsync -a%sxHAX %s %s", verboseFlag, mountPoint, destination))
	if err != nil {
		return fmt.Errorf("Failed to sync image contents to %s: %s", destination, err)
	}

	unmountResult, err := pmt.UnMountImage(mountPoint, false)
	if err != nil {
		return fmt.Errorf("Error when unmounting image at %s: %s", mountPoint, err)
	}
	if !unmountResult {
		return fmt.Errorf("Could not unmount image at %s", mountPoint)
	}

	return nil
}
