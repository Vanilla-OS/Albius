package main

import (
	"os"

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/albius/core"
)

func main() {
	if reexec.Init() { // needed for subprocesses
		panic("Failed to initialize reexec")
	}

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
