package ffi

/*
   #include "lib/ffi_types.h"
   #include "lib/ffi_funcs.h"
*/
import "C"

import (
	"fmt"
	"regexp"
	"os"
)

// TODO: Create, remove, resize partitions

//export NewPartition
func NewPartition(target *C.disk, name, fsType *C.char, start, end C.int) {
	createPartCmd := "parted -s %s mkpart%s \"%s\" %s %d %d"

	var partType string
	if C.GoString(target._label) == "msdos" {
		partType = " primary"
	} else {
		partType = ""
	}

	err := RunCommand(fmt.Sprintf(createPartCmd, C.GoString(target._path), partType, C.GoString(name), C.GoString(fsType), start, end))
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create partition: %s", err)
		C._ffi_println(C.CString(errorMsg))
		os.Exit(1)
	}
}

//export RemovePartition
func RemovePartition(target *C.partition) {
	rmPartCmd := "parted -s %s rm %s"

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(C.GoString(target._path))
	part := partExpr.FindString(C.GoString(target._path))

	err := RunCommand(fmt.Sprintf(rmPartCmd, disk, part))
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to remove partition: %s", err)
		C._ffi_println(C.CString(errorMsg))
		os.Exit(1)
	}
}

//export ResizePartition
func ResizePartition(target *C.partition, newEnd C.int) {
	resizePartCmd := "parted -s %s resizepart %s %d"

	diskExpr := regexp.MustCompile("^/dev/[a-zA-Z]+([0-9]+[a-z][0-9]+)?")
	partExpr := regexp.MustCompile("[0-9]+$")
	disk := diskExpr.FindString(C.GoString(target._path))
	part := partExpr.FindString(C.GoString(target._path))

	err := RunCommand(fmt.Sprintf(resizePartCmd, disk, part, newEnd))
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to resize partition: %s", err)
		C._ffi_println(C.CString(errorMsg))
		os.Exit(1)
	}
}