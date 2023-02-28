#ifndef FFI_TYPES
#define FFI_TYPES

#include <stdio.h>
#include <stdlib.h>

#include <string.h>

typedef struct _partition {
    char *_path;
    int _number;
    char *_start;
    char *_end;
    char *_size;
    char *_type;
    char *_filesystem;
} partition;

typedef struct _disk {
    char *_path;
    char *_size;
    char *_model;
    char *_transport;
    int _logical_sector_size;
    int _physical_sector_size;
    char *_label;
    int _max_partitions;
    partition *_partitions;
    int _partitions_count;
} disk;

#endif
