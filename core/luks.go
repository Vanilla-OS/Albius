package albius

import (
	"fmt"
	"os/exec"
)

func IsLuks(part *Partition) (bool, error) {
	isLuksCmd := "cryptsetup isLuks %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(isLuksCmd, part.Path))
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}

func LuksOpen(part *Partition, mapping string) error {
	luksOpenCmd := "cryptsetup open %s %s"

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
