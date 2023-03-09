package albius

import (
	"encoding/json"
	"fmt"
	"os"
)

type Recipe struct {
	setup Setup
	mountpoints []Mountpoint
}

type Setup struct {
	steps []Step
}

type Step struct {
	disk, operation string
	params []interface{}
}

type Mountpoint struct {
	partition, target string
}

func ReadRecipe(path string) (*Recipe, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read recipe: %s", err)
	}

	var recipe *Recipe
	err = json.Unmarshal(content, &recipe)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal recipe: %s", err)
	}

	return recipe, nil
}