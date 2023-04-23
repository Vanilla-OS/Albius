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

// LuksOpen opens a LUKS-encrypted partition, mapping the unencrypted filesystem
// to /dev/mapper/<mapping>.
// If password is an empty string, cryptsetup will prompt the password when
// executed.
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
