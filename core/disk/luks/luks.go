package disk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vanilla-os/albius/core/util"
)

type Partition interface {
	GetUUID() (string, error)
	GetPath() string
}

func IsLuks(part Partition) (bool, error) {
	isLuksCmd := "cryptsetup isLuks %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(isLuksCmd, part.GetPath()))
	err := cmd.Run()
	if err != nil {
		// We expect the command to return exit status 1 if partition isn't LUKS-encrypted
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			} else {
				return false, fmt.Errorf("failed to check if %s is LUKS-encrypted: %s", part.GetPath(), string(exitError.Stderr))
			}
		}
		return false, fmt.Errorf("failed to check if %s is LUKS-encrypted: %s", part.GetPath(), err)
	}

	return true, nil
}

// LuksOpen opens a LUKS-encrypted partition, mapping the unencrypted filesystem
// to /dev/mapper/<mapping>.
//
// If password is an empty string, cryptsetup will prompt the password when
// executed.
//
// WARNING: This function will return an error if mapping already exists, use
// LuksTryOpen() to open a device while ignoring existing mappings
func LuksOpen(part Partition, mapping, password string) error {
	var luksOpenCmd string
	if password != "" {
		luksOpenCmd = fmt.Sprintf("echo '%s' | ", password)
	} else {
		luksOpenCmd = ""
	}

	luksOpenCmd += "cryptsetup open %s %s"

	err := util.RunCommand(fmt.Sprintf(luksOpenCmd, part.GetPath(), mapping))
	if err != nil {
		return fmt.Errorf("failed to open LUKS-encrypted partition: %s", err)
	}

	return nil
}

// LuksTryOpen opens a LUKS-encrypted partition, failing silently if mapping
// already exists.
//
// This is useful for when we pass a mapping like "luks-<uuid>", which we are
// certain is unique and the operation failing means that the device is already
// open.
//
// The function still returns other errors, however.
func LuksTryOpen(part Partition, mapping, password string) error {
	_, err := os.Stat(fmt.Sprintf("/dev/mapper/%s", mapping))
	if err == nil { // Mapping exists, do nothing
		return nil
	} else if os.IsNotExist(err) {
		return LuksOpen(part, mapping, password)
	} else {
		return fmt.Errorf("failed to try-open LUKS-encrypted partition: %s", err)
	}
}

func LuksClose(mapping string) error {
	luksCloseCmd := "cryptsetup close %s"

	err := util.RunCommand(fmt.Sprintf(luksCloseCmd, mapping))
	if err != nil {
		return fmt.Errorf("failed to close LUKS-encrypted partition: %s", err)
	}

	return nil
}

func LuksFormat(part Partition, password string) error {
	luksFormatCmd := "echo '%s' | cryptsetup -q luksFormat %s"

	err := util.RunCommand(fmt.Sprintf(luksFormatCmd, password, part.GetPath()))
	if err != nil {
		return fmt.Errorf("failed to create LUKS-encrypted partition: %s", err)
	}

	return nil
}

func GenCrypttab(targetRoot string, entries [][]string) error {
	file, err := os.Create(fmt.Sprintf("%s/etc/crypttab", targetRoot))
	if err != nil {
		return err
	}

	defer file.Close()

	for _, entry := range entries {
		fmtEntry := strings.Join(entry, " ")
		_, err := file.Write(append([]byte(fmtEntry), '\n'))
		if err != nil {
			return err
		}
	}

	return nil
}

func GetLUKSFilesystemByPath(path string) (string, error) {
	lsblkCmd := "lsblk -n -o FSTYPE %s | sed '/crypto_LUKS/d'"
	output, err := util.OutputCommand(fmt.Sprintf(lsblkCmd, path))
	if err != nil {
		return "", fmt.Errorf("failed to get encrypted partition FSTYPE: %s", err)
	}

	return output, nil
}
