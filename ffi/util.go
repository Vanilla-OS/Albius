package ffi

import "C"
import "unsafe"

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
