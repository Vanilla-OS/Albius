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
	"os"
	"os/exec"
	"strings"

	"github.com/vanilla-os/albius/native"
)

type LocateDiskOutput struct {
	Disk native.Disk
}

//export LocateDisk
func LocateDisk(diskname *C.char) *C.disk {
	findPartitionCmd := "parted -sj %s print | sed -r 's/^(\\s*)\"(.)/\\1\"\\U\\2/g' | sed -r 's/(\\S)-(\\S)/\\1\\U\\2/g'"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(findPartitionCmd, C.GoString(diskname)))
	output, err := cmd.Output()
	if err != nil {
		C._ffi_println(C.CString("Failed to run command."))
		os.Exit(1)
	}

	var decoded LocateDiskOutput
	dec := json.NewDecoder(strings.NewReader(string(output)))
	err = dec.Decode(&decoded)
	if err != nil {
		C._ffi_println(C.CString("Failed to retrieve partition."))
		return nil
	}

	device := decoded.Disk

	return DiskToCStruct(device)
}

//export Mount
func Mount(part *C.partition, location *C.char) {
	mountCmd := "mount -m %s %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(mountCmd, C.GoString(part._path), C.GoString(location)))
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

	cmd := exec.Command("sh", "-c", fmt.Sprintf(umountCmd, C.GoString(part._path)))
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
