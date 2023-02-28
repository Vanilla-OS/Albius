package main

import (
	"github.com/vanilla-os/albius/albius"
)
import "C"

type Setup func()

var setup Setup = ffi.Start

func main() {}
