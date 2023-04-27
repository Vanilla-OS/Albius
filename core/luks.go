package albius

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func IsLuks(part *Partition) (bool, error) {
	isLuksCmd := "cryptsetup isLuks %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(isLuksCmd, part.Path))
	err := cmd.Run()
	if err != nil {
		// We expect the command to return exit status 1 if partition isn't
		// LUKS-encrypted
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("Failed to check if %s is LUKS-encrypted: %s", part.Path, err)
	}

	return true, nil
}

func IsPathLuks(path string) (bool, error) {
	dummyPartition := Partition{}
	dummyPartition.Path = path

	return IsLuks(&dummyPartition)
}

// LuksOpen opens a LUKS-encrypted partition, mapping the unencrypted filesystem
// to /dev/mapper/<mapping>.
// If password is an empty string, cryptsetup will prompt the password when
// executed.
// WARNING: This function will return an error if mapping already exists, use
// LuksTryOpen() to open a device while ignoring existing mappings
func LuksOpen(part *Partition, mapping, password string) error {
	var luksOpenCmd string
	if password != "" {
		luksOpenCmd = fmt.Sprintf("echo '%s' | ", password)
	} else {
		luksOpenCmd = ""
	}

	luksOpenCmd += "cryptsetup open %s %s"

	err := RunCommand(fmt.Sprintf(luksOpenCmd, part.Path, mapping))
	if err != nil {
		return fmt.Errorf("Failed to open LUKS-encrypted partition: %s", err)
	}

	return nil
}

// LuksTryOpen opens a LUKS-encrypted partition, failing silently if mapping
// already exists.
// This is useful for when we pass a mapping like "luks-<uuid>", which we are
// certain is unique and the operation failing means that the device is already
// open.
// The function still returns other errors, however.
func LuksTryOpen(part *Partition, mapping, password string) error {
	_, err := os.Stat(fmt.Sprintf("/dev/mapper/%s", mapping))
	if err == nil { // Mapping exists, do nothing
		return nil
	} else if os.IsNotExist(err) {
		return LuksOpen(part, mapping, password)
	} else {
		return fmt.Errorf("Failed to try-open LUKS-encrypted partition: %s", err)
	}
}

func LuksClose(mapping string) error {
	luksCloseCmd := "cryptsetup close %s"

	err := RunCommand(fmt.Sprintf(luksCloseCmd, mapping))
	if err != nil {
		return fmt.Errorf("Failed to close LUKS-encrypted partition: %s", err)
	}

	return nil
}

func LuksFormat(part *Partition, password string) error {
	luksFormatCmd := "echo '%s' | cryptsetup -q luksFormat %s"

	err := RunCommand(fmt.Sprintf(luksFormatCmd, password, part.Path))
	if err != nil {
		return fmt.Errorf("Failed to create LUKS-encrypted partition: %s", err)
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
	lsblkCmd := "lsblk -d -n -o FSTYPE %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(lsblkCmd, path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get encrypted partition FSTYPE: %s", err)
	}

	return string(output[:len(output)-1]), nil
}
