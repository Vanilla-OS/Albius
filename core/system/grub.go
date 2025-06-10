package system

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/albius/core/util"
)

type GrubConfig map[string]string

const (
	BIOS = "i386-pc"
	EFI  = "x86_64-efi"
)

type FirmwareType string

func GetGrubConfig(targetRoot string) (GrubConfig, error) {
	targetRootGrubFile := filepath.Join(targetRoot, "/etc/default/grub")

	// If grub config file doesn't exist yet, return an empty map
	if _, err := os.Stat(targetRootGrubFile); os.IsNotExist(err) {
		return GrubConfig{}, nil
	}

	content, err := os.ReadFile(targetRootGrubFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read GRUB config file: %s", err)
	}

	config := GrubConfig{}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line != "" && line[0] != '#' {
			kv := strings.SplitN(line, "=", 2)
			config[kv[0]] = kv[1]
		}
	}

	return config, nil
}

func WriteGrubConfig(targetRoot string, config GrubConfig) error {
	fileContents := []byte{}
	for k, v := range config {
		line := fmt.Sprintf("%s=%s\n", k, v)
		fileContents = append(fileContents, []byte(line)...)
	}

	targetRootGrubFile := filepath.Join(targetRoot, "/etc/default/grub")
	err := os.WriteFile(targetRootGrubFile, fileContents, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write GRUB config file: %s", err)
	}

	return nil
}

func AddGrubScript(targetRoot, scriptPath string) error {
	// Ensure script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("error adding GRUB script: %s does not exist", scriptPath)
	}

	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read GRUB script at %s: %s", scriptPath, err)
	}

	targetRootPath := filepath.Join(targetRoot, "/etc/grub.d", filepath.Base(scriptPath))
	err = os.WriteFile(targetRootPath, contents, 0o755) // Grub expects script to be executable
	if err != nil {
		return fmt.Errorf("failed to writing GRUB script to %s: %s", targetRootPath, err)
	}

	return nil
}

func RemoveGrubScript(targetRoot, scriptName string) error {
	targetRootPath := filepath.Join(targetRoot, "/etc/grub.d", scriptName)

	// Ensure script exists
	if _, err := os.Stat(targetRootPath); os.IsNotExist(err) {
		return fmt.Errorf("error removing GRUB script: %s does not exist", targetRootPath)
	}

	err := os.Remove(targetRootPath)
	if err != nil {
		return fmt.Errorf("error removing GRUB script: %s", err)
	}

	return nil
}

func RunGrubInstall(targetRoot, bootDirectory, diskPath string, target FirmwareType, entryName string, removable bool, efiDevice ...string) error {
	// Mount necessary targets for chroot
	if targetRoot != "" {
		requiredBinds := []string{"/dev", "/dev/pts", "/proc", "/sys", "/run"}
		for _, bind := range requiredBinds {
			targetBind := filepath.Join(targetRoot, bind)
			err := util.RunCommand(fmt.Sprintf("mount --bind %s %s", bind, targetBind))
			if err != nil {
				return fmt.Errorf("failed to mount %s to %s: %s", bind, targetRoot, err)
			}
		}
	}

	grubInstallCmd := "grub-install --no-nvram %s --bootloader-id=%s --boot-directory %s --target=%s --uefi-secure-boot %s"

	removableStr := ""
	if removable {
		removableStr = "--removable"
	}

	command := fmt.Sprintf(grubInstallCmd, removableStr, entryName, bootDirectory, target, diskPath)

	var err error
	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, command)
	} else {
		err = util.RunCommand(command)
	}
	if err != nil {
		return fmt.Errorf("failed to run grub-install: %s", err)
	}

	if targetRoot != "" {
		return nil
	}

	if !removable && target == EFI {
		efibootmgrCmd := "efibootmgr --create --disk=%s --part=%s --label=%s --loader=\"\\EFI\\%s\\shimx64.efi\""
		if len(efiDevice) == 0 || efiDevice[0] == "" {
			return errors.New("EFI device was not specified")
		}
		diskName, part := util.SeparateDiskPart(efiDevice[0])
		err = util.RunCommand(fmt.Sprintf(efibootmgrCmd, diskName, part, entryName, entryName))
		if err != nil {
			return fmt.Errorf("failed to run efibootmgr for grub-install: %s", err)
		}
	}

	return nil
}

func RunGrubMkconfig(targetRoot, output string) error {
	grubMkconfigCmd := "grub-mkconfig -o %s"

	var err error
	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, fmt.Sprintf(grubMkconfigCmd, output))
	} else {
		err = util.RunCommand(fmt.Sprintf(grubMkconfigCmd, output))
	}
	if err != nil {
		return err
	}

	return nil
}
