#ifndef FFI_FUNCS
#define FFI_FUNCS

#include <stdio.h>
#include <string.h>
#include "ffi_types.h"

void _ffi_println(char *s);
void add_path_to_partitions(partition *parts, int count, char *path);

#endif
