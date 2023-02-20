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

//export LocateDisk
func LocateDisk(diskname *C.char) *C.disk {
	findPartitionCmd := "lsblk -nJ %s | sed 's/maj:min/majmin/g' | sed -r 's/^(\\s*)\"(.)/\\1\"\\U\\2/g'"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(findPartitionCmd, C.GoString(diskname)))
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

//export Mount
func Mount(part *C.partition, location *C.char) {
	mountCmd := "mount -m %s %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(mountCmd, "/dev/"+C.GoString(part.name), C.GoString(location)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		C._ffi_println(C.CString("Failed to run command"))
		os.Exit(1)
	}
}

//export UmountPartition
func UmountPartition(part *C.partition) {
	umountCmd := "umount %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(umountCmd, "/dev/"+C.GoString(part.name)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		C._ffi_println(C.CString("Failed to run command"))
		os.Exit(1)
	}
}

//export UmountDirectory
func UmountDirectory(dir *C.char) {
	umountCmd := "umount %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(umountCmd, C.GoString(dir)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		C._ffi_println(C.CString("Failed to run command"))
		os.Exit(1)
	}
}
