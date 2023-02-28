package ffi

/*
   #include "lib/ffi_types.h"
   #include "lib/ffi_funcs.h"
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
)

//export Unsquashfs
func Unsquashfs(filesystem, destination *C.char, force C.int) {
	unsquashfsCmd := "unsquashfs%s -d %s"

	var forceFlag string
	if force == 1 {
		forceFlag = " -f"
	} else {
		forceFlag = ""
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf(unsquashfsCmd, forceFlag, C.GoString(destination)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		C._ffi_println(C.CString("Failed to run command"))
		os.Exit(1)
	}
}
