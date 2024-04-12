package util

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// RunCommand executes a command in a subshell
//
// envVars are environement variables in the form MYVAR=myvalue that will be passed to the command
func RunCommand(command string, envVars ...string) error {
	stderr := new(bytes.Buffer)

	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		return errors.New(stderr.String())
	}

	return nil
}

// OutputCommand executes a command in a subshell and returns its output
func OutputCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return strings.TrimSpace(string(out)), errors.New(string(exitErr.Stderr))
		}
		return strings.TrimSpace(string(out)), err
	}

	return strings.TrimSpace(string(out)), err
}

// RunInChroot executes a command in a subshell while chrooted into the specified root
func RunInChroot(root, command string) error {
	cmd := exec.Command("chroot", root, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return errors.New(string(exitErr.Stderr))
		}
		return err
	}

	return nil
}

// SeparateDiskPart receives a path (e.g. /dev/sda1) and separates it into
// the device root and partition number
func SeparateDiskPart(path string) (string, string) {
	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(path)
	part := partExpr.FindString(path)

	return disk, part
}
