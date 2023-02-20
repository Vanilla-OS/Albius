package ffi

// #cgo CFLAGS: -I..
/*
   #include "ffi_types.h"

   static void _ffi_println(char *s) {
       printf("%s\n", s);
   }
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"github.com/vanilla-os/albius/native"
	"os"
	"os/exec"
)

type LocateDiskOutput struct {
	Blockdevices []native.Blockdevice
}

var FindPartitionCmd = "lsblk -nJ %s | sed 's/maj:min/majmin/g' | sed -r 's/^(\\s*)\"(.)/\\1\"\\U\\2/g'"

//export LocateDisk
func LocateDisk(diskname *C.char) *C.disk {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(FindPartitionCmd, C.GoString(diskname)))
	output, err := cmd.Output()
	if err != nil {
		C._ffi_println(C.CString("Failed to run command"))
		os.Exit(1)
	}

	var devices LocateDiskOutput
    json.Unmarshal(output, &devices)

	if len(devices.Blockdevices) == 1 {
		return BlockdeviceToCStruct(devices.Blockdevices[0])
	}

	return nil
}
