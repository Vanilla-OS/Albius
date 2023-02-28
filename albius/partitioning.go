package ffi

/*
   #include "lib/ffi_types.h"
   #include "lib/ffi_funcs.h"
*/
import "C"

import (
	"fmt"
	"os"
)

// TODO: Create, remove, resize partitions

//export NewPartition
func NewPartition(target *C.disk, fsType *C.char, start, end C.int) {
	createPartCmd := "parted -s %s mkpart primary %s %s %d %d"

	err := RunCommand(fmt.Sprintf(createPartCmd, C.GoString(target._path), C.GoString(fsType), start, end))
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create partition: %s", err)
		C._ffi_println(C.CString(errorMsg))
		os.Exit(1)
	}
}