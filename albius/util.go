package ffi

import "C"

import (
	"os"
	"os/exec"
	"unsafe"
)

func BoolToCInt(boolval bool) C.int {
	if boolval {
		return 1
	}

	return 0
}

func StringListToCArray(stringList []string) (**C.char, C.size_t) {
	cArray := C.malloc(C.size_t(len(stringList)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	goArray := (*[1<<30 - 1]*C.char)(cArray)

	for i, val := range stringList {
		goArray[i] = C.CString(val)
	}

	return (**C.char)(cArray), C.size_t(len(stringList))
}

func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
