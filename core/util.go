package albius

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
)

// RunCommand executes a command in a subshell.
func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if err != nil && ok {
		return errors.New(string(exitErr.Stderr))
	} else if err != nil {
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
