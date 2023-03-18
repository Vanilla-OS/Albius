package main

import (
	"os"

	"github.com/vanilla-os/albius/core"
)

func main() {
	recipe, err := albius.ReadRecipe(os.Args[1])
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
}
