package albius

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Recipe struct {
	Setup       Setup
	Mountpoints []Mountpoint
}

type Setup struct {
	Steps []Step
}

type Step struct {
	Disk, Operation string
	Params          []interface{}
}

type Mountpoint struct {
	Partition, Target string
}

func ReadRecipe(path string) (*Recipe, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read recipe: %s", err)
	}

	var recipe Recipe
	dec := json.NewDecoder(strings.NewReader(string(content)))
	dec.DisallowUnknownFields()
	dec.UseNumber()
	err = dec.Decode(&recipe)
	if err != nil {
		return nil, fmt.Errorf("Failed to read recipe: %s", err)
	}

	// Convert json.Number to int64
	for i := 0; i < len(recipe.Setup.Steps); i++ {
		step := &recipe.Setup.Steps[i]
		formattedParams := []interface{}{}
		for _, param := range step.Params {
			var dummy json.Number
			dummy = "1"
			if reflect.TypeOf(param) == reflect.TypeOf(dummy) {
				convertedParam, err := param.(json.Number).Int64()
				if err != nil {
					return nil, fmt.Errorf("Failed to convert recipe parameter: %s", err)
				}
				formattedParams = append(formattedParams, convertedParam)
			} else {
				formattedParams = append(formattedParams, param)
			}
		}
		step.Params = formattedParams
	}

	return &recipe, nil
}
