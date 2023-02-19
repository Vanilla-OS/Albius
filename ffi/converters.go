package ffi

// #cgo CFLAGS: -I..
// #include "ffi_types.h"
import "C"

import (
	"github.com/vanilla-os/albius/native"
	"unsafe"
)

func BlockdeviceToCStruct(block native.Blockdevice) C.disk {
	cPart := C.disk{}

	cPart.name = C.CString(block.Name)
	cPart.majmin = C.CString(block.Majmin)
	cPart.fssize = C.CString(block.Fssize)
	cPart.pttype = C.CString(block.Pttype)
	cPart.rm = BoolToCInt(block.Rm)
	cPart.ro = BoolToCInt(block.Ro)
	cPart.mountpoints, cPart.mountpoints_size = StringListToCArray(block.Mountpoints)
	cPart.partitions, cPart.partitions_size = PartitionSliceToCArray(block.Children)

	return cPart
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
