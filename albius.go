package main

import (
	"github.com/vanilla-os/albius/ffi"
)

type Setup func()

var setup Setup = ffi.Start

func main() {}
