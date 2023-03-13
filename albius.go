package main

import (
	"github.com/vanilla-os/albius/core"
)

func main() {
	recipe, err := albius.ReadRecipe("recipe_template.json")
	if err != nil {
		panic(err)
	}

	err = recipe.RunSetup()
	if err != nil {
		panic(err)
	}

	err = recipe.SetupMountpoints()
	if err != nil {
		panic(err)
	}

	err = recipe.Install()
	if err != nil {
		panic(err)
	}

	err = recipe.RunPostInstall()
	if err != nil {
		panic(err)
	}

	// TODO: Call abroot-adapter
}
