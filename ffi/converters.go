package ffi

// #cgo CFLAGS: -I..
// #include "ffi_types.h"
import "C"

import (
	"github.com/vanilla-os/albius/native"
	"unsafe"
)

func BlockdeviceToCStruct(block native.Blockdevice) *C.disk {
	cPart := C.malloc(C.size_t(unsafe.Sizeof(C.sizeof_disk)))

	(*C.disk)(cPart).name = C.CString(block.Name)
	(*C.disk)(cPart).majmin = C.CString(block.Majmin)
	(*C.disk)(cPart).fssize = C.CString(block.Fssize)
	(*C.disk)(cPart).pttype = C.CString(block.Pttype)
	(*C.disk)(cPart).rm = BoolToCInt(block.Rm)
	(*C.disk)(cPart).ro = BoolToCInt(block.Ro)
	(*C.disk)(cPart).mountpoints, (*C.disk)(cPart).mountpoints_size = StringListToCArray(block.Mountpoints)
	(*C.disk)(cPart).partitions, (*C.disk)(cPart).partitions_size = PartitionSliceToCArray(block.Children)

	return (*C.disk)(cPart)
}

func PartitionToCStruct(part native.Partition) C.partition {
	cPart := C.partition{}

	// HACK: For some reason, the first time we convert using C.CString returns garbage
	_ = C.CString(part.Name)

	cPart.name = C.CString(part.Name)
	cPart.majmin = C.CString(part.Majmin)
	cPart.fssize = C.CString(part.Fssize)
	cPart.fstype = C.CString(part.Fstype)
	cPart.rm = BoolToCInt(part.Rm)
	cPart.ro = BoolToCInt(part.Ro)
	cPart.mountpoints, cPart.mountpoints_size = StringListToCArray(part.Mountpoints)

	return cPart
}

func PartitionSliceToCArray(slice []native.Partition) (*C.partition, C.size_t) {
	cArray := C.malloc(C.size_t(len(slice)) * C.size_t(unsafe.Sizeof(C.sizeof_partition)))
	goArray := (*[1<<30 - 1]C.partition)(cArray)

	for i, val := range slice {
		goArray[i] = PartitionToCStruct(val)
	}

	return (*C.partition)(cArray), C.size_t(len(slice))
}
