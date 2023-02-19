#ifndef FFI_TYPES
#define FFI_TYPES

#include <stdio.h>
#include <stdlib.h>

typedef struct _partition {
    char *name;
    char *majmin;
    int rm;
    char *fssize;
    char *fstype;
    int ro;
    char **mountpoints;
    size_t mountpoints_size;
} partition;

typedef struct _disk {
    char *name;
    char *majmin;
    char *fssize;
    char *pttype;
    int rm;
    int ro;
    char **mountpoints;
    size_t mountpoints_size;
    partition *partitions;
    size_t partitions_size;
} disk;

#endif
