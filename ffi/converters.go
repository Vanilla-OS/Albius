package ffi

// #cgo CFLAGS: -I..
/*
   #include <string.h>
   #include "ffi_types.h"

   void add_path_to_partitions(partition *parts, int count, char *path) {
      for(int i = 0; i < count; i++) {
          char *path_prefix;
          char num_str[3];

          if (atoi(&path[strlen(path)-1]) == 0)
              path_prefix = strdup(path);
          else {
              path_prefix = (char *)malloc((strlen(path) + 1) * sizeof(char));
              strcpy(path_prefix, path);
              strcat(path_prefix, "p");
          }

          sprintf(num_str, "%d", parts[i]._number);

          parts[i]._path = (char *)malloc((strlen(path_prefix) + 3) * sizeof(char));
          strcpy(parts[i]._path, path_prefix);
          strcat(parts[i]._path, num_str);

          free(path_prefix);
      }
   }
*/
import "C"

import (
	"github.com/vanilla-os/albius/native"
	"unsafe"
)

func DiskToCStruct(block native.Disk) *C.disk {
	cPart := C.malloc(C.size_t(unsafe.Sizeof(C.sizeof_disk)))

	(*C.disk)(cPart)._path = C.CString(block.Path)
	(*C.disk)(cPart)._size = C.CString(block.Size)
	(*C.disk)(cPart)._model = C.CString(block.Model)
	(*C.disk)(cPart)._transport = C.CString(block.Transport)
	(*C.disk)(cPart)._logical_sector_size = (C.int)(block.LogicalSectorSize)
	(*C.disk)(cPart)._physical_sector_size = (C.int)(block.PhysicalSectorSize)
	(*C.disk)(cPart)._label = C.CString(block.Label)
	(*C.disk)(cPart)._max_partitions = (C.int)(block.MaxPartitions)
	(*C.disk)(cPart)._partitions, (*C.disk)(cPart)._partitions_count = PartitionSliceToCArray(block.Partitions)

	C.add_path_to_partitions((*C.disk)(cPart)._partitions, (*C.disk)(cPart)._partitions_count, (*C.disk)(cPart)._path)

	return (*C.disk)(cPart)
}

func PartitionToCStruct(part native.Partition) C.partition {
	cPart := C.partition{}

    // HACK: The first call to `C.CString()` returns garbage
    _ = C.CString(part.Start)

	cPart._number = (C.int)(part.Number)
	cPart._start = C.CString(part.Start)
	cPart._end = C.CString(part.End)
	cPart._size = C.CString(part.Size)
	cPart._type = C.CString(part.Type)
	cPart._filesystem = C.CString(part.Filesystem)

	return cPart
}

func PartitionSliceToCArray(slice []native.Partition) (*C.partition, C.int) {
	cArray := C.malloc(C.size_t(len(slice)) * C.size_t(unsafe.Sizeof(C.sizeof_partition)))
	goArray := (*[1<<30 - 1]C.partition)(cArray)

	for i, val := range slice {
		goArray[i] = PartitionToCStruct(val)
	}

	return (*C.partition)(cArray), (C.int)(len(slice))
}
