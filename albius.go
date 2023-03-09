package main

import (
	"fmt"

	"github.com/vanilla-os/albius/core"
)

func main() {
	disk, err := albius.LocateDisk("/dev/nvme0n1")
	if err != nil {
		panic(err)
	}

	fmt.Println(disk)
}
