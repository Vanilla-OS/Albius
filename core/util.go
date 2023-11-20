package albius

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
)

// RunCommand executes a command in a subshell
func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
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

func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	var convertedVal reflect.Value
	if structFieldType != val.Type() {
		// Type conversions
		if structFieldType.Kind() == reflect.Int && val.Type().Kind() == reflect.Float64 {
			convertedVal = reflect.ValueOf(int(val.Interface().(float64)))
		} else if structFieldType.Name() == "DiskLabel" && val.Type().Kind() == reflect.String {
			convertedVal = reflect.ValueOf(DiskLabel(val.Interface().(string)))
		} else {
			return fmt.Errorf("provided value type for %s did not match obj field type. Expected %v, got %v", name, structFieldType, val.Type())
		}
	} else {
		convertedVal = val
	}

	structFieldValue.Set(convertedVal)
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
