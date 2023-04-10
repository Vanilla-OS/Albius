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