#ifndef LVM_CGO_H
#define LVM_CGO_H

#include <stdlib.h>
#include <string.h>
#include <lvm2cmd.h>

void lvm_log_capture_fn(int level, const char *file, int line,
                        int dm_errno, const char *format);

char **log_output();

#endif