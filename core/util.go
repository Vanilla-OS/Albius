package albius

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
)

func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
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
			return fmt.Errorf("Provided value type for %s did not match obj field type. Expected %v, got %v.", name, structFieldType, val.Type())
		}
	} else {
		convertedVal = val
	}

	structFieldValue.Set(convertedVal)
	return nil
}
